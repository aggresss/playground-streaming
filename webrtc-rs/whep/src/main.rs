use std::collections::HashMap;
use std::fs::File;
use std::future::Future;
use std::io::BufReader;
use std::net::SocketAddr;
use std::path::Path;
use std::pin::Pin;
use std::sync::{Arc, Mutex};
use std::time::Duration;

use anyhow::Result;
use clap::Parser;
use http_body_util::{BodyExt, Full};
use hyper::body::Bytes;
use hyper::server::conn::http1;
use hyper::service::Service;
use hyper::{body::Incoming as IncomingBody, Method, Request, Response, StatusCode};
use hyper_util::rt::TokioIo;
use tokio::net::TcpListener;

use tokio::sync::Notify;
use webrtc::api::interceptor_registry::register_default_interceptors;
use webrtc::api::media_engine::{MediaEngine, MIME_TYPE_H264, MIME_TYPE_OPUS};
use webrtc::api::APIBuilder;
use webrtc::ice_transport::ice_connection_state::RTCIceConnectionState;
use webrtc::ice_transport::ice_server::RTCIceServer;
use webrtc::interceptor::registry::Registry;
use webrtc::media::io::h264_reader::H264Reader;
use webrtc::media::io::ogg_reader::OggReader;
use webrtc::media::Sample;
use webrtc::peer_connection::configuration::RTCConfiguration;
use webrtc::peer_connection::peer_connection_state::RTCPeerConnectionState;
use webrtc::peer_connection::sdp::session_description::RTCSessionDescription;
use webrtc::peer_connection::RTCPeerConnection;
use webrtc::rtp_transceiver::rtp_codec::RTCRtpCodecCapability;
use webrtc::track::track_local::track_local_static_sample::TrackLocalStaticSample;
use webrtc::track::track_local::TrackLocal;
use webrtc::Error;

#[derive(Parser)]
#[command(version, about, long_about = None)]
#[derive(Debug, Clone)]
struct WhepHandler {
    #[arg(short, long, default_value = "0.0.0.0:8082")]
    listen_addr: String,

    #[arg(short, long, default_value = "127.0.0.1", env = "CANDIDATES")]
    candidates: Vec<String>,

    #[arg(short = 'u', long, default_value_t = 15060)]
    ice_udp_port: u16,

    #[arg(short = 't', long, default_value_t = 15060)]
    ice_tcp_port: u16,

    #[arg(short, long, default_value = "output.ogg")]
    audio_file_name: String,

    #[arg(short, long, default_value = "output.h264")]
    video_file_name: String,

    #[arg(short = 'p', long, default_value = "20")]
    ogg_page_ms: u64,

    #[arg(short = 'f', long, default_value = "41")]
    h264_frame_ms: u64,

    #[arg(skip)]
    whep_clients: HashMap<String, Arc<RTCPeerConnection>>,
}

impl WhepHandler {
    fn init(&mut self) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
        let video_path = self.video_file_name.as_str();
        if !Path::new(video_path).exists() {
            return Err(Error::new(format!("video file: '{video_path}' not exist")).into());
        }
        let audio_path = self.audio_file_name.as_str();
        if !Path::new(audio_path).exists() {
            return Err(Error::new(format!("audio file: '{audio_path}' not exist")).into());
        }
        Ok(())
    }

    async fn create_whep_client(
        &self,
        path: &str,
        offer_str: &str,
    ) -> Result<String, Box<dyn std::error::Error + Send + Sync>> {
        let mut m = MediaEngine::default();
        m.register_default_codecs()?;

        let mut registry = Registry::new();
        registry = register_default_interceptors(registry, &mut m)?;

        let api = APIBuilder::new()
            .with_media_engine(m)
            .with_interceptor_registry(registry)
            .build();

        let config = RTCConfiguration {
            ice_servers: vec![RTCIceServer {
                urls: vec!["stun:stun.l.google.com:19302".to_owned()],
                ..Default::default()
            }],
            ..Default::default()
        };

        let peer_connection = Arc::new(api.new_peer_connection(config).await?);

        let notify_tx = Arc::new(Notify::new());
        let notify_video = notify_tx.clone();
        let notify_audio = notify_tx.clone();

        let (done_tx, mut done_rx) = tokio::sync::mpsc::channel::<()>(1);
        let video_done_tx = done_tx.clone();
        let audio_done_tx = done_tx.clone();

        // Create a video track
        let video_track = Arc::new(TrackLocalStaticSample::new(
            RTCRtpCodecCapability {
                mime_type: MIME_TYPE_H264.to_owned(),
                ..Default::default()
            },
            "video".to_owned(),
            "webrtc-rs".to_owned(),
        ));

        let rtp_sender = peer_connection
            .add_track(Arc::clone(&video_track) as Arc<dyn TrackLocal + Send + Sync>)
            .await?;

        tokio::spawn(async move {
            let mut rtcp_buf = vec![0u8; 1500];
            while let Ok((_, _)) = rtp_sender.read(&mut rtcp_buf).await {}
        });

        let video_file_name = self.video_file_name.clone();
        let video_file_interval = self.h264_frame_ms;
        tokio::spawn(async move {
            let file = File::open(&video_file_name)?;
            let reader = BufReader::new(file);
            let mut h264 = H264Reader::new(reader, 1_048_576);

            notify_video.notified().await;

            println!("play video from disk file {video_file_name}");

            let mut ticker = tokio::time::interval(Duration::from_millis(video_file_interval));
            loop {
                let nal = match h264.next_nal() {
                    Ok(nal) => nal,
                    Err(err) => {
                        println!("All video frames parsed and sent: {err}");
                        break;
                    }
                };

                video_track
                    .write_sample(&Sample {
                        data: nal.data.freeze(),
                        duration: Duration::from_secs(1),
                        ..Default::default()
                    })
                    .await?;

                let _ = ticker.tick().await;
            }

            let _ = video_done_tx.try_send(());

            Result::<()>::Ok(())
        });

        // Create a audio track
        let audio_track = Arc::new(TrackLocalStaticSample::new(
            RTCRtpCodecCapability {
                mime_type: MIME_TYPE_OPUS.to_owned(),
                ..Default::default()
            },
            "audio".to_owned(),
            "webrtc-rs".to_owned(),
        ));

        let rtp_sender = peer_connection
            .add_track(Arc::clone(&audio_track) as Arc<dyn TrackLocal + Send + Sync>)
            .await?;

        tokio::spawn(async move {
            let mut rtcp_buf = vec![0u8; 1500];
            while let Ok((_, _)) = rtp_sender.read(&mut rtcp_buf).await {}
            Result::<()>::Ok(())
        });

        let audio_file_name = self.audio_file_name.clone();
        let audio_file_interval = self.ogg_page_ms;
        tokio::spawn(async move {
            let file = File::open(audio_file_name)?;
            let reader = BufReader::new(file);
            let (mut ogg, _) = OggReader::new(reader, true)?;

            // Wait for connection established
            notify_audio.notified().await;

            println!("play audio from disk file output.ogg");

            let mut ticker = tokio::time::interval(Duration::from_millis(audio_file_interval));

            let mut last_granule: u64 = 0;
            while let Ok((page_data, page_header)) = ogg.parse_next_page() {
                let sample_count = page_header.granule_position - last_granule;
                last_granule = page_header.granule_position;
                let sample_duration = Duration::from_millis(sample_count * 1000 / 48000);

                audio_track
                    .write_sample(&Sample {
                        data: page_data.freeze(),
                        duration: sample_duration,
                        ..Default::default()
                    })
                    .await?;

                let _ = ticker.tick().await;
            }

            let _ = audio_done_tx.try_send(());

            Result::<()>::Ok(())
        });

        peer_connection.on_ice_connection_state_change(Box::new(
            move |connection_state: RTCIceConnectionState| {
                println!("Connection State has changed {connection_state}");
                if connection_state == RTCIceConnectionState::Connected {
                    notify_tx.notify_waiters();
                }
                Box::pin(async {})
            },
        ));

        peer_connection.on_peer_connection_state_change(Box::new(
            move |s: RTCPeerConnectionState| {
                println!("Peer Connection State has changed: {s}");

                if s == RTCPeerConnectionState::Failed {
                    println!("Peer Connection has gone to failed exiting");
                    let _ = done_tx.try_send(());
                }

                Box::pin(async {})
            },
        ));

        let offer = RTCSessionDescription::offer(offer_str.to_owned())?;
        peer_connection.set_remote_description(offer).await?;
        let answer = peer_connection.create_answer(None).await?;
        let mut gather_complete = peer_connection.gathering_complete_promise().await;
        peer_connection.set_local_description(answer).await?;
        let _ = gather_complete.recv().await;

        match peer_connection.local_description().await {
            Some(local_desc) => {
                // self.whep_clients.insert(String::from(path), peer_connection.clone());
                return Ok(local_desc.sdp.into())
            }
            _ => {
                return Err(Error::new(format!("generate answer failed")).into());
            }
        }
    }

    fn delete_whep_client(
        &mut self,
        path: &str,
    ) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
        Ok(())
    }
}

#[derive(Debug, Clone)]
struct Svc {
    whep: Arc<WhepHandler>,
}

impl Service<Request<IncomingBody>> for Svc {
    type Response = Response<Full<Bytes>>;
    type Error = hyper::Error;
    type Future = Pin<Box<dyn Future<Output = Result<Self::Response, Self::Error>> + Send>>;

    fn call(&self, req: Request<IncomingBody>) -> Self::Future {
        let mut builder = Response::builder();
        let path = String::from(req.uri().path());

        if req.headers().contains_key("Origin") {
            builder = builder.header(
                "Access-Control-Allow-Origin",
                req.headers()["Origin"].to_str().unwrap(),
            );
            builder = builder.header("Access-Control-Allow-Credentials", "true");
        }

        match req.method() {
            &Method::POST => {
                let svc = self.clone();
                return Box::pin(async move {
                    let offer_str = String::from_utf8(req.collect().await?.to_bytes().to_vec()).unwrap();
                    let answer_str = svc.whep.create_whep_client(path.as_str(), offer_str.as_str()).await.unwrap();
                    Ok(builder
                        .header("Content-Type", "application/sdp")
                        .status(StatusCode::CREATED)
                        .body(Full::new(Bytes::from(answer_str)))
                        .unwrap())
                });
            }
            &Method::DELETE => {
                return Box::pin(async {
                    Ok(builder
                        .status(StatusCode::NOT_FOUND)
                        .body(Full::new(Bytes::from("")))
                        .unwrap())
                });
            }
            &Method::OPTIONS => {
                if req.headers().contains_key("Access-Control-Request-Method") {
                    builder = builder.header(
                        "Access-Control-Allow-Methods",
                        req.headers()["Access-Control-Request-Method"]
                            .to_str()
                            .unwrap(),
                    )
                }
                if req.headers().contains_key("Access-Control-Request-Headers") {
                    builder = builder.header(
                        "Access-Control-Allow-Headers",
                        req.headers()["Access-Control-Request-Headers"]
                            .to_str()
                            .unwrap(),
                    )
                }
                return Box::pin(async {
                    Ok(builder
                        .status(StatusCode::NO_CONTENT)
                        .body(Full::new(Bytes::from("")))
                        .unwrap())
                });
            }
            _ => {
                return Box::pin(async {
                    Ok(builder
                        .status(StatusCode::NOT_FOUND)
                        .body(Full::new(Bytes::from("")))
                        .unwrap())
                });
            }
        };
    }
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let mut whep_handler = WhepHandler::parse();
    let addr: SocketAddr = whep_handler
        .listen_addr
        .clone()
        .parse()
        .expect("Unable to parse socket address");
    let listener = TcpListener::bind(addr).await?;
    println!("Listening on http://{}", addr);

    whep_handler.init()?;

    let svc = Svc {
        whep: Arc::new(whep_handler),
    };

    loop {
        let (stream, _) = listener.accept().await?;
        let io = TokioIo::new(stream);

        let svc_clone = svc.clone();
        tokio::task::spawn(async move {
            if let Err(err) = http1::Builder::new().serve_connection(io, svc_clone).await {
                println!("Error serving connection: {:?}", err);
            }
        });
    }
}

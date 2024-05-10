use std::future::Future;
use std::net::SocketAddr;
use std::pin::Pin;
use std::sync::{Arc, Mutex};
use std::time::Duration;


use http_body_util::Full;
use hyper::body::Bytes;
use hyper::server::conn::http1;
use hyper::service::Service;
use hyper::{body::Incoming as IncomingBody, Method, Request, Response, StatusCode};
use hyper_util::rt::TokioIo;
use tokio::net::TcpListener;

static HTTP_ADDR: &str = "0.0.0.0:8082";
static CANDIDATE: &str = "127.0.0.1";
static ICE_UDP_PORT: u16 = 15060;
static ICE_TCP_PORT: u16 = 15060;
static AUDIO_FILE_NAME: &str = "output.ogg";
static VIDEO_FILE_NAME: &str = "output.h264";
static OGG_PAGE_DURATION: Duration = Duration::from_millis(20);
static H264_FRAME_DURATION: Duration = Duration::from_millis(41);

#[derive(Debug, Clone)]
struct WhepHandler {
    // http_addr: String,
    // candidates: Vec<String>,
    // ice_udp_port: u16,
    // ice_tcp_port: u16,
    // audio_file_name: String,
    // video_file_name: String,
    // ogg_page_duration: Duration,
    // h264_frame_duration: Duration,
}

#[derive(Debug, Clone)]
struct Svc {
    whep: Arc<Mutex<WhepHandler>>,
}

impl Service<Request<IncomingBody>> for Svc {
    type Response = Response<Full<Bytes>>;
    type Error = hyper::Error;
    type Future = Pin<Box<dyn Future<Output = Result<Self::Response, Self::Error>> + Send>>;

    fn call(&self, req: Request<IncomingBody>) -> Self::Future {
        let mut builder = Response::builder();

        if req.headers().contains_key("Origin") {
            builder = builder.header(
                "Access-Control-Allow-Origin",
                req.headers()["Origin"].to_str().unwrap(),
            );
            builder = builder.header("Access-Control-Allow-Credentials", "true");
        }

        let res = match req.method() {
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
                Ok(builder
                    .status(StatusCode::NO_CONTENT)
                    .body(Full::new(Bytes::from("")))
                    .unwrap())
            }
            _ => Ok(builder
                .status(StatusCode::NOT_FOUND)
                .body(Full::new(Bytes::from("")))
                .unwrap()),
        };

        Box::pin(async { res })
    }
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let addr: SocketAddr = "0.0.0.0:8082"
        .parse()
        .expect("Unable to parse socket address");
    let listener = TcpListener::bind(addr).await?;
    println!("Listening on http://{}", addr);

    let svc = Svc {
        whep: Arc::new(Mutex::new(WhepHandler{})),
    };
    loop {
        let (stream, _) = listener.accept().await?;
        let io = TokioIo::new(stream);

        let svc_clone = svc.clone();
        tokio::task::spawn(async move {
            if let Err(err) = http1::Builder::new()
                .serve_connection(io, svc_clone)
                .await
            {
                println!("Error serving connection: {:?}", err);
            }
        });
    }
}

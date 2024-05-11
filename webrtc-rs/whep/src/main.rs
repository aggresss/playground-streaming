use std::collections::HashMap;
use std::future::Future;
use std::net::SocketAddr;
use std::pin::Pin;
use std::sync::{Arc, Mutex};
use std::time::Duration;

use clap::Parser;
use http_body_util::Full;
use hyper::body::Bytes;
use hyper::server::conn::http1;
use hyper::service::Service;
use hyper::{body::Incoming as IncomingBody, Method, Request, Response, StatusCode};
use hyper_util::rt::TokioIo;
use tokio::net::TcpListener;

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
    ogg_page_ms: usize,

    #[arg(short = 'f', long, default_value = "41")]
    h264_frame_ms: usize,

    #[arg(skip)]
    whep_clients: HashMap<String, String>,
}

impl WhepHandler {
    fn create_whep_client(
        &mut self,
        path: &str,
        offer: &str,
    ) -> Result<String, Box<dyn std::error::Error>> {
        Ok("".into())
    }

    fn delete_whep_client(&mut self, path: &str) -> Result<(), Box<dyn std::error::Error>> {
        Ok(())
    }
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
    let whep_handler = WhepHandler::parse();
    let addr: SocketAddr = whep_handler
        .listen_addr
        .clone()
        .parse()
        .expect("Unable to parse socket address");
    let listener = TcpListener::bind(addr).await?;
    println!("Listening on http://{}", addr);

    let svc = Svc {
        whep: Arc::new(Mutex::new(whep_handler)),
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

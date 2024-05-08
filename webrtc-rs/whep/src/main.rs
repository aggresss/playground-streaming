use std::net::SocketAddr;

use http_body_util::{combinators::BoxBody, BodyExt, Empty, Full};
use hyper::body::Bytes;
use hyper::server::conn::http1;
use hyper::service::service_fn;
use hyper::{Method, Request, Response, StatusCode};
use hyper_util::rt::TokioIo;
use tokio::net::TcpListener;

async fn handle_whep(
    req: Request<hyper::body::Incoming>,
) -> Result<Response<BoxBody<Bytes, hyper::Error>>, hyper::Error> {
    let mut builder = Response::builder();
    if req.headers().contains_key("Origin") {
        builder = builder.header("Access-Control-Allow-Origin", req.headers()["Origin"].to_str().unwrap());
        builder = builder.header("Access-Control-Allow-Credentials",  "true");
    }

    match req.method() {
        &Method::POST => {
            Ok(Response::new(req.into_body().boxed()))
        }
        &Method::DELETE => {
            Ok(Response::new(req.into_body().boxed()))
        }
        &Method::OPTIONS => {
            if req.headers().contains_key("Access-Control-Request-Method") {
                builder = builder.header("Access-Control-Allow-Methods", req.headers()["Access-Control-Request-Method"].to_str().unwrap())
            }
            if req.headers().contains_key("Access-Control-Request-Headers") {
                builder = builder.header("Access-Control-Allow-Headers", req.headers()["Access-Control-Request-Headers"].to_str().unwrap())
            }
            Ok(builder.status(StatusCode::NO_CONTENT).body(empty()).unwrap())
        }
        _ => {
            Ok(builder.status(StatusCode::NOT_FOUND).body(empty()).unwrap())
        }
    }
}

fn empty() -> BoxBody<Bytes, hyper::Error> {
    Empty::<Bytes>::new()
        .map_err(|never| match never {})
        .boxed()
}

fn full<T: Into<Bytes>>(chunk: T) -> BoxBody<Bytes, hyper::Error> {
    Full::new(chunk.into())
        .map_err(|never| match never {})
        .boxed()
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
    let addr = SocketAddr::from(([0, 0, 0, 0], 8082));

    let listener = TcpListener::bind(addr).await?;
    println!("Listening on http://{}", addr);
    loop {
        let (stream, _) = listener.accept().await?;
        let io = TokioIo::new(stream);

        tokio::task::spawn(async move {
            if let Err(err) = http1::Builder::new()
                .serve_connection(io, service_fn(handle_whep))
                .await
            {
                println!("Error serving connection: {:?}", err);
            }
        });
    }
}

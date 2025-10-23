use actix_web::{web, App, HttpResponse, HttpServer, Responder};
use serde::{Deserialize, Serialize};
use std::env;

#[derive(Debug, Serialize, Deserialize)]
struct WeatherTweet {
    municipality: String,
    temperature: i32,
    humidity: i32,
    weather: String,
}

#[derive(Debug, Serialize)]
struct Response {
    status: String,
}

async fn health_check() -> impl Responder {
    HttpResponse::Ok().json(Response {
        status: "OK".to_string(),
    })
}

async fn receive_tweet(tweet: web::Json<WeatherTweet>) -> impl Responder {
    let go_service_url = env::var("GO_SERVICE_URL")
        .unwrap_or_else(|_| "http://go-grpc-client:8080/tweet".to_string());

    log::info!(
        "Received tweet: municipality={}, temp={}, humidity={}, weather={}",
        tweet.municipality,
        tweet.temperature,
        tweet.humidity,
        tweet.weather
    );

    // Forward to Go service
    let client = reqwest::Client::new();
    match client
        .post(&go_service_url)
        .json(&tweet.into_inner())
        .send()
        .await
    {
        Ok(response) => {
            if response.status().is_success() {
                log::info!("Successfully forwarded to Go service");
                HttpResponse::Ok().json(Response {
                    status: "Tweet received and forwarded".to_string(),
                })
            } else {
                log::error!("Go service returned error: {}", response.status());
                HttpResponse::InternalServerError().json(Response {
                    status: "Error forwarding to Go service".to_string(),
                })
            }
        }
        Err(e) => {
            log::error!("Failed to forward to Go service: {}", e);
            HttpResponse::InternalServerError().json(Response {
                status: format!("Error: {}", e),
            })
        }
    }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    env_logger::init_from_env(env_logger::Env::new().default_filter_or("info"));

    let port = env::var("PORT").unwrap_or_else(|_| "8000".to_string());
    let bind_addr = format!("0.0.0.0:{}", port);

    log::info!("Starting Rust API server on {}", bind_addr);

    HttpServer::new(|| {
        App::new()
            .route("/health", web::get().to(health_check))
            .route("/tweet", web::post().to(receive_tweet))
    })
    .bind(&bind_addr)?
    .workers(4)
    .run()
    .await
}
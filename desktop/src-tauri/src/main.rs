#![cfg_attr(
    all(not(debug_assertions), target_os = "windows"),
    windows_subsystem = "windows"
)]

use std::process::{Child, Command};
use std::sync::Mutex;
use tauri::Manager;

struct ServerProcess(Mutex<Option<Child>>);

fn main() {
    tauri::Builder::default()
        .manage(ServerProcess(Mutex::new(None)))
        .setup(|app| {
            let app_handle = app.handle();

            // Start the embedded Go server
            #[cfg(target_os = "windows")]
            let server_path = app_handle
                .path_resolver()
                .resolve_resource("binaries/server.exe")
                .expect("failed to resolve server binary");

            #[cfg(not(target_os = "windows"))]
            let server_path = app_handle
                .path_resolver()
                .resolve_resource("binaries/server")
                .expect("failed to resolve server binary");

            let server_process = Command::new(&server_path)
                .env("PORT", "8081")
                .env("ENV", "desktop")
                .env("DATABASE_URL", "sqlite://data.db")
                .env("REDIS_URL", "memory://")
                .spawn();

            match server_process {
                Ok(child) => {
                    let state: tauri::State<ServerProcess> = app.state();
                    *state.0.lock().unwrap() = Some(child);
                    println!("Server started successfully");
                }
                Err(e) => {
                    eprintln!("Failed to start server: {}", e);
                    // Continue without local server - will use cloud backend
                }
            }

            Ok(())
        })
        .on_window_event(|event| {
            if let tauri::WindowEvent::CloseRequested { .. } = event.event() {
                // Stop the server when window closes
                let app = event.window().app_handle();
                let state: tauri::State<ServerProcess> = app.state();
                if let Some(mut child) = state.0.lock().unwrap().take() {
                    let _ = child.kill();
                }
            }
        })
        .invoke_handler(tauri::generate_handler![get_server_status])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}

#[tauri::command]
fn get_server_status(state: tauri::State<ServerProcess>) -> bool {
    state.0.lock().unwrap().is_some()
}

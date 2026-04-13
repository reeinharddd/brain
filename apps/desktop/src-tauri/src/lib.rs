use serde::{Deserialize, Serialize};

const AUTH_SERVICE: &str = "brain-desktop";
const AUTH_ACCOUNT: &str = "auth-session";

#[derive(Debug, Serialize, Deserialize)]
struct DesktopAuthSession {
  token: String,
  refresh_token: Option<String>,
  expires_at: Option<String>,
  refresh_expires_at: Option<String>,
}

#[tauri::command]
fn save_auth_session(session_json: String) -> Result<(), String> {
  let entry = keyring::Entry::new(AUTH_SERVICE, AUTH_ACCOUNT)
    .map_err(|err| format!("create keychain entry: {err}"))?;
  entry
    .set_password(&session_json)
    .map_err(|err| format!("save auth session: {err}"))
}

#[tauri::command]
fn load_auth_session() -> Result<Option<String>, String> {
  let entry = keyring::Entry::new(AUTH_SERVICE, AUTH_ACCOUNT)
    .map_err(|err| format!("create keychain entry: {err}"))?;
  match entry.get_password() {
    Ok(value) => Ok(Some(value)),
    Err(err) => {
      let message = err.to_string();
      if message.contains("not found") {
        Ok(None)
      } else {
        Err(format!("load auth session: {err}"))
      }
    }
  }
}

#[tauri::command]
fn clear_auth_session() -> Result<(), String> {
  let entry = keyring::Entry::new(AUTH_SERVICE, AUTH_ACCOUNT)
    .map_err(|err| format!("create keychain entry: {err}"))?;
  match entry.delete_password() {
    Ok(()) => Ok(()),
    Err(err) => {
      let message = err.to_string();
      if message.contains("not found") {
        Ok(())
      } else {
        Err(format!("clear auth session: {err}"))
      }
    }
  }
}

#[tauri::command]
fn open_external_url(url: String) -> Result<(), String> {
  let url = url.trim();
  if url.is_empty() {
    return Err("url is empty".to_string());
  }

  let result = if cfg!(target_os = "windows") {
    std::process::Command::new("cmd")
      .args(["/C", "start", "", url])
      .spawn()
  } else if cfg!(target_os = "macos") {
    std::process::Command::new("open").arg(url).spawn()
  } else {
    std::process::Command::new("xdg-open").arg(url).spawn()
  };

  result.map(|_| ()).map_err(|err| format!("open url: {err}"))
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
  tauri::Builder::default()
    .invoke_handler(tauri::generate_handler![
      save_auth_session,
      load_auth_session,
      clear_auth_session,
      open_external_url,
    ])
    .setup(|app| {
      if cfg!(debug_assertions) {
        app.handle().plugin(
          tauri_plugin_log::Builder::default()
            .level(log::LevelFilter::Info)
            .build(),
        )?;
      }
      Ok(())
    })
    .run(tauri::generate_context!())
    .expect("error while running tauri application");
}

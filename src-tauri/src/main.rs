use tauri::api::process::{Command, CommandEvent};
use serde::{Deserialize, Serialize};
use tokio::time::{timeout, Duration};

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct AnalyzeInputRequest {
    file_name: String,
    file_path: String,
    flatten_threshold: Option<i32>,
    array_item_name: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct AnalysisMetadata {
    total_fields: usize,
    array_fields: usize,
    optional_fields: usize,
    table_count: usize,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct AnalysisTreeNode {
    id: String,
    name: String,
    path: String,
    #[serde(rename = "type")]
    node_type: String,
    optional: bool,
    is_array: bool,
    child_count: usize,
    #[serde(default)]
    children: Vec<AnalysisTreeNode>,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct RecordChildGroup {
    name: String,
    count: usize,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    key_values: Vec<String>,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct RecordInstance {
    attributes: std::collections::HashMap<String, String>,
    #[serde(default, skip_serializing_if = "Vec::is_empty")]
    children: Vec<RecordChildGroup>,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct RecordGroup {
    type_name: String,
    instances: Vec<RecordInstance>,
}

#[derive(Debug, Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct AnalyzeInputResponse {
    schema_tree: Vec<AnalysisTreeNode>,
    sql: String,
    metadata: AnalysisMetadata,
    #[serde(default)]
    records: Vec<RecordGroup>,
}

#[tauri::command]
async fn analyze_structured_data(
    request: AnalyzeInputRequest,
) -> Result<AnalyzeInputResponse, String> {
    validate_file_name(&request.file_name)?;

    let payload = serde_json::to_vec(&request)
        .map_err(|error| format!("Failed to encode sidecar request: {error}"))?;
    let response = timeout(Duration::from_secs(60), run_sidecar(payload))
        .await
        .map_err(|_| String::from("StructLens engine timed out while analyzing the file."))??;

    serde_json::from_str(&response)
        .map_err(|error| format!("Failed to decode backend response: {error}"))
}

fn validate_file_name(file_name: &str) -> Result<(), String> {
    let extension = std::path::Path::new(file_name)
        .extension()
        .and_then(|ext| ext.to_str())
        .map(|ext| ext.to_ascii_lowercase())
        .ok_or_else(|| String::from("Unsupported file type. Supported formats: .json, .xml"))?;

    if extension != "json" && extension != "xml" {
        return Err(String::from(
            "Unsupported file type. Supported formats: .json, .xml",
        ));
    }

    Ok(())
}

async fn run_sidecar(payload: Vec<u8>) -> Result<String, String> {
    let (mut receiver, mut child) = Command::new_sidecar("structlens-engine")
        .map_err(|error| format!("Failed to prepare StructLens engine sidecar: {error}"))?
        .spawn()
        .map_err(|error| format!("Failed to launch StructLens engine: {error}"))?;

    child
        .write(&payload)
        .map_err(|error| format!("Failed to send request to StructLens engine: {error}"))?;
    child
        .write(b"\n")
        .map_err(|error| format!("Failed to finalize engine request: {error}"))?;

    let mut stdout = String::new();
    let mut stderr = String::new();
    let mut exit_code = None;

    while let Some(event) = receiver.recv().await {
        match event {
            CommandEvent::Stdout(line) => stdout.push_str(&line),
            CommandEvent::Stderr(line) => {
                stderr.push_str(&line);
                stderr.push('\n');
            }
            CommandEvent::Error(message) => stderr.push_str(&message),
            CommandEvent::Terminated(payload) => {
                exit_code = payload.code;
                break;
            }
            _ => {}
        }
    }

    if exit_code == Some(0) {
        return Ok(stdout);
    }

    if stderr.trim().is_empty() {
        return Err(String::from(
            "StructLens engine failed without an error message.",
        ));
    }

    Err(stderr.trim().to_string())
}

fn main() {
    tauri::Builder::default()
        .invoke_handler(tauri::generate_handler![analyze_structured_data])
        .run(tauri::generate_context!())
        .expect("error while running StructLens application");
}

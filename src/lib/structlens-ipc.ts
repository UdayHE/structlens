import { invoke } from "@tauri-apps/api/tauri";
import {
  type AnalyzeInputRequest,
  type AnalyzeInputResponse,
} from "@/types/app";

export async function analyzeStructuredData(
  request: AnalyzeInputRequest,
): Promise<AnalyzeInputResponse> {
  return invoke<AnalyzeInputResponse>("analyze_structured_data", { request });
}

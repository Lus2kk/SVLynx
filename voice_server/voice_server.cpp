#include "httplib.h"
#include "whisper.h"
#include <filesystem>
#include <fstream>
#include <iostream>
#include <string>
#include <chrono>
#include <random>
#include <sstream>
#include <vector>

namespace fs = std::filesystem;

whisper_context* g_whisper_ctx = nullptr;

std::string generateUUID() {
    std::random_device rd;
    std::mt19937 gen(rd());
    std::uniform_int_distribution<> dis(0, 15);
    std::uniform_int_distribution<> dis2(8, 11);

    std::stringstream ss;
    ss << std::hex;
    for (int i = 0; i < 8; i++) ss << dis(gen);
    ss << "-";
    for (int i = 0; i < 4; i++) ss << dis(gen);
    ss << "-4";
    for (int i = 0; i < 3; i++) ss << dis(gen);
    ss << "-";
    ss << dis2(gen);
    for (int i = 0; i < 3; i++) ss << dis(gen);
    ss << "-";
    for (int i = 0; i < 12; i++) ss << dis(gen);
    return ss.str();
}

std::string getDatePath() {
    auto now = std::chrono::system_clock::now();
    auto time = std::chrono::system_clock::to_time_t(now);
    std::tm tm = *std::localtime(&time);
    char buf[20];
    std::strftime(buf, sizeof(buf), "%Y-%m-%d", &tm);
    return std::string(buf);
}

std::string escapeJson(const std::string& s) {
    std::string out;
    for (char c : s) {
        if (c == '"') out += "\\\"";
        else if (c == '\\') out += "\\\\";
        else if (c == '\n') out += "\\n";
        else if (c == '\r') out += "\\r";
        else if (c == '\t') out += "\\t";
        else out += c;
    }
    return out;
}

std::string transcribeAudio(const std::string& filePath) {
    if (!g_whisper_ctx) return "";

    std::string wavPath = filePath + ".wav";
    std::string cmd = "ffmpeg -y -i \"" + filePath + "\" -ar 16000 -ac 1 -f wav \"" + wavPath + "\" 2>/dev/null";
    int ret = system(cmd.c_str());
    if (ret != 0 || !fs::exists(wavPath)) return "";

    std::ifstream f(wavPath, std::ios::binary);
    if (!f) { fs::remove(wavPath); return ""; }

    f.seekg(44);
    std::vector<int16_t> pcm16(
        (std::istreambuf_iterator<char>(f)),
        std::istreambuf_iterator<char>()
    );
    f.close();
    fs::remove(wavPath);

    if (pcm16.empty()) return "";

    std::vector<float> pcmf32(pcm16.size());
    for (size_t i = 0; i < pcm16.size(); i++)
        pcmf32[i] = pcm16[i] / 32768.0f;

    whisper_full_params params = whisper_full_default_params(WHISPER_SAMPLING_GREEDY);
    params.language    = "ru";
    params.translate   = false;
    params.print_progress = false;
    params.print_realtime = false;
    params.print_timestamps = false;

    if (whisper_full(g_whisper_ctx, params, pcmf32.data(), (int)pcmf32.size()) != 0)
        return "";

    std::string result;
    int n = whisper_full_n_segments(g_whisper_ctx);
    for (int i = 0; i < n; i++)
        result += whisper_full_get_segment_text(g_whisper_ctx, i);

    if (!result.empty() && result[0] == ' ')
        result = result.substr(1);

    return result;
}

int main() {
    whisper_context_params cparams = whisper_context_default_params();
    g_whisper_ctx = whisper_init_from_file_with_params("models/ggml-base.bin", cparams);
    if (!g_whisper_ctx) {
        std::cerr << "Warning: failed to load Whisper model, transcription disabled" << std::endl;
    } else {
        std::cout << "Whisper model loaded" << std::endl;
    }

    httplib::Server svr;

    const std::string UPLOAD_DIR = "./uploads/voice";
    const std::string BASE_URL = "https://svlynx.site";
    const std::string GO_URL = "http://localhost:8080";

    fs::create_directories(UPLOAD_DIR);

    svr.set_pre_routing_handler([](const httplib::Request& req, httplib::Response& res) {
        res.set_header("Access-Control-Allow-Origin", "*");
        res.set_header("Access-Control-Allow-Methods", "GET, POST, OPTIONS");
        res.set_header("Access-Control-Allow-Headers", "Content-Type, Authorization");
        if (req.method == "OPTIONS") {
            res.status = 204;
            return httplib::Server::HandlerResponse::Handled;
        }
        return httplib::Server::HandlerResponse::Unhandled;
    });

    svr.Get(R"(/uploads/voice/(\d{4}-\d{2}-\d{2})/(.+))", [&](const httplib::Request& req, httplib::Response& res) {
        std::string date = req.matches[1];
        std::string file = req.matches[2];
        std::string path = UPLOAD_DIR + "/" + date + "/" + file;

        if (!fs::exists(path)) {
            res.status = 404;
            res.set_content("Not found", "text/plain");
            return;
        }

        std::ifstream f(path, std::ios::binary);
        std::string content((std::istreambuf_iterator<char>(f)),
                             std::istreambuf_iterator<char>());

        std::string contentType = "audio/webm";
        if (file.find(".mp4") != std::string::npos) contentType = "audio/mp4";
        else if (file.find(".ogg") != std::string::npos) contentType = "audio/ogg";
        res.set_content(content, contentType);
    });

    svr.Post("/voice/upload", [&](const httplib::Request& req, httplib::Response& res) {
        std::string chatId, senderId, recipientId;
        std::string fileContent;
        std::string fileExt = ".webm";

        if (!req.is_multipart_form_data()) {
            res.status = 400;
            res.set_content(R"({"error":"expected multipart/form-data"})", "application/json");
            return;
        }

        auto chatIt = req.form.fields.find("chat_id");
        if (chatIt != req.form.fields.end()) chatId = chatIt->second.content;

        auto senderIt = req.form.fields.find("sender_id");
        if (senderIt != req.form.fields.end()) senderId = senderIt->second.content;

        auto recipIt = req.form.fields.find("recipient_id");
        if (recipIt != req.form.fields.end()) recipientId = recipIt->second.content;

        std::string durationStr = "0";
        auto durIt = req.form.fields.find("duration");
        if (durIt != req.form.fields.end()) durationStr = durIt->second.content;

        auto fileIt = req.form.files.find("file");
        if (fileIt != req.form.files.end()) {
            fileContent = fileIt->second.content;
            if (fileIt->second.content_type.find("ogg") != std::string::npos)
                fileExt = ".ogg";
            else if (fileIt->second.content_type.find("mp4") != std::string::npos)
                fileExt = ".mp4";
        }

        if (fileContent.empty() || chatId.empty() || senderId.empty() || recipientId.empty()) {
            res.status = 400;
            res.set_content(R"({"error":"missing required fields"})", "application/json");
            return;
        }

        std::string datePath = getDatePath();
        std::string dirPath = UPLOAD_DIR + "/" + datePath;
        fs::create_directories(dirPath);

        std::string uuid = generateUUID();
        std::string filename = uuid + fileExt;
        std::string filePath = dirPath + "/" + filename;

        std::ofstream outFile(filePath, std::ios::binary);
        outFile.write(fileContent.data(), fileContent.size());
        outFile.close();

        std::string audioUrl = BASE_URL + "/uploads/voice/" + datePath + "/" + filename;

        std::string transcript = transcribeAudio(filePath);
        std::cout << "Transcript: " << transcript << std::endl;

        httplib::Client goClient(GO_URL);

        std::string body = R"({"chat_id":")" + chatId +
                   R"(","sender_id":")" + senderId +
                   R"(","recipient_id":")" + recipientId +
                   R"(","audio_url":")" + audioUrl +
                   R"(","duration":)" + durationStr +
                   R"(,"transcript":")" + escapeJson(transcript) + R"("})";

        httplib::Headers headers = {
            {"Content-Type", "application/json"},
            {"Authorization", req.get_header_value("Authorization")}
        };

        auto goRes = goClient.Post("/chat/messages/voice", headers, body, "application/json");

        if (!goRes || goRes->status != 201) {
            res.status = 500;
            res.set_content(R"({"error":"failed to create message in db"})", "application/json");
            return;
        }

        res.set_content(goRes->body, "application/json");
        std::cout << "Voice saved: " << filePath << std::endl;
    });
        
    svr.Post("/voice/upload/channel", [&](const httplib::Request& req, httplib::Response& res) {
        if (!req.is_multipart_form_data()) {
            res.status = 400;
            res.set_content(R"({"error":"expected multipart/form-data"})", "application/json");
            return;
        }

        std::string senderId;
        auto senderIt = req.form.fields.find("sender_id");
        if (senderIt != req.form.fields.end()) senderId = senderIt->second.content;

        std::string durationStr = "0";
        auto durIt = req.form.fields.find("duration");
        if (durIt != req.form.fields.end()) durationStr = durIt->second.content;

        std::string fileContent;
        std::string fileExt = ".webm";
        auto fileIt = req.form.files.find("file");
        if (fileIt != req.form.files.end()) {
            fileContent = fileIt->second.content;
            if (fileIt->second.content_type.find("ogg") != std::string::npos)
                fileExt = ".ogg";
            else if (fileIt->second.content_type.find("mp4") != std::string::npos)
                fileExt = ".mp4";
        }

        if (fileContent.empty() || senderId.empty()) {
            res.status = 400;
            res.set_content(R"({"error":"missing required fields"})", "application/json");
            return;
        }

        std::string datePath = getDatePath();
        std::string dirPath = UPLOAD_DIR + "/" + datePath;
        fs::create_directories(dirPath);

        std::string uuid = generateUUID();
        std::string filename = uuid + fileExt;
        std::string filePath = dirPath + "/" + filename;

        std::ofstream outFile(filePath, std::ios::binary);
        outFile.write(fileContent.data(), fileContent.size());
        outFile.close();

        std::string audioUrl = BASE_URL + "/uploads/voice/" + datePath + "/" + filename;

        std::string transcript = transcribeAudio(filePath);

        std::string responseBody = R"({"url":")" + audioUrl +
                                   R"(","duration":)" + durationStr +
                                   R"(,"transcript":")" + escapeJson(transcript) + R"("})";

        res.set_content(responseBody, "application/json");

        std::cout << "Channel voice saved: " << filePath << std::endl;
    });

    std::cout << "Voice server running on port 9090" << std::endl;
    svr.listen("0.0.0.0", 9090);
    return 0;
}
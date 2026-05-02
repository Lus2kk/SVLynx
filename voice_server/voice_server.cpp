#include "httplib.h"
#include <filesystem>
#include <fstream>
#include <iostream>
#include <string>
#include <chrono>
#include <random>
#include <sstream>

namespace fs = std::filesystem;

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

int main() {
    httplib::Server svr;

    const std::string UPLOAD_DIR = "./uploads/voice";
    const std::string BASE_URL = "http://localhost:9090";
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
        res.set_content(content, "audio/webm");
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

        httplib::Client goClient(GO_URL);

        std::string body = R"({"chat_id":")" + chatId +
                           R"(","sender_id":")" + senderId +
                           R"(","recipient_id":")" + recipientId +
                           R"(","audio_url":")" + audioUrl + R"("})";

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

    std::cout << "Voice server running on port 9090" << std::endl;
    svr.listen("0.0.0.0", 9090);
    return 0;
}
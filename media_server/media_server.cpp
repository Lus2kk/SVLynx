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

// Определяем тип медиа и расширение по content-type
struct MediaInfo {
    std::string ext;        // расширение файла (.jpg, .mp4, ...)
    std::string subDir;     // подпапка (images, videos, audio, files)
    std::string msgType;    // тип сообщения для Go (image, video, audio, file)
    std::string mimeType;   // MIME для отдачи файла
};

MediaInfo detectMediaInfo(const std::string& contentType) {
    // Images
    if (contentType.find("image/jpeg") != std::string::npos)  return {".jpg",  "images", "image", "image/jpeg"};
    if (contentType.find("image/png")  != std::string::npos)  return {".png",  "images", "image", "image/png"};
    if (contentType.find("image/gif")  != std::string::npos)  return {".gif",  "images", "image", "image/gif"};
    if (contentType.find("image/webp") != std::string::npos)  return {".webp", "images", "image", "image/webp"};
    // Videos
    if (contentType.find("video/mp4")  != std::string::npos)  return {".mp4",  "videos", "video", "video/mp4"};
    if (contentType.find("video/webm") != std::string::npos)  return {".webm", "videos", "video", "video/webm"};
    if (contentType.find("video/quicktime") != std::string::npos) return {".mov", "videos", "video", "video/quicktime"};
    // Audio
    if (contentType.find("audio/mpeg") != std::string::npos)  return {".mp3",  "audio",  "audio", "audio/mpeg"};
    if (contentType.find("audio/ogg")  != std::string::npos)  return {".ogg",  "audio",  "audio", "audio/ogg"};
    if (contentType.find("audio/wav")  != std::string::npos)  return {".wav",  "audio",  "audio", "audio/wav"};
    // Documents
    if (contentType.find("application/pdf") != std::string::npos) return {".pdf", "files", "file", "application/pdf"};
    if (contentType.find("application/zip") != std::string::npos) return {".zip", "files", "file", "application/zip"};
    if (contentType.find("text/plain")      != std::string::npos) return {".txt", "files", "file", "text/plain"};
    // Fallback
    return {".bin", "files", "file", "application/octet-stream"};
}

// Экранируем строку для JSON
std::string jsonEscape(const std::string& s) {
    std::string result;
    for (char c : s) {
        if (c == '"')  result += "\\\"";
        else if (c == '\\') result += "\\\\";
        else if (c == '\n') result += "\\n";
        else if (c == '\r') result += "\\r";
        else result += c;
    }
    return result;
}

int main() {
    httplib::Server svr;

    const std::string UPLOAD_DIR = "./uploads/media";
    const std::string BASE_URL = "http://localhost:9091";
    const std::string GO_URL     = "http://localhost:8080";
    const size_t      MAX_SIZE   = 50 * 1024 * 1024; // 50 MB

    fs::create_directories(UPLOAD_DIR);

    // CORS — как у Кости
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

    // Отдача файлов: GET /uploads/media/{subdir}/{date}/{filename}
    svr.Get(R"(/uploads/media/([a-z]+)/(\d{4}-\d{2}-\d{2})/(.+))",
        [&](const httplib::Request& req, httplib::Response& res) {
            std::string subDir   = req.matches[1];
            std::string date     = req.matches[2];
            std::string filename = req.matches[3];
            std::string path     = UPLOAD_DIR + "/" + subDir + "/" + date + "/" + filename;

            if (!fs::exists(path)) {
                res.status = 404;
                res.set_content("Not found", "text/plain");
                return;
            }

            std::ifstream f(path, std::ios::binary);
            std::string content((std::istreambuf_iterator<char>(f)),
                                 std::istreambuf_iterator<char>());

            // Определяем MIME по расширению
            std::string mime = "application/octet-stream";
            if (filename.find(".jpg")  != std::string::npos) mime = "image/jpeg";
            else if (filename.find(".png")  != std::string::npos) mime = "image/png";
            else if (filename.find(".gif")  != std::string::npos) mime = "image/gif";
            else if (filename.find(".webp") != std::string::npos) mime = "image/webp";
            else if (filename.find(".mp4")  != std::string::npos) mime = "video/mp4";
            else if (filename.find(".webm") != std::string::npos) mime = "video/webm";
            else if (filename.find(".mp3")  != std::string::npos) mime = "audio/mpeg";
            else if (filename.find(".ogg")  != std::string::npos) mime = "audio/ogg";
            else if (filename.find(".pdf")  != std::string::npos) mime = "application/pdf";

            res.set_content(content, mime);
        }
    );

    // Загрузка файла: POST /media/upload
    svr.Post("/media/upload", [&](const httplib::Request& req, httplib::Response& res) {

        if (!req.is_multipart_form_data()) {
            res.status = 400;
            res.set_content(R"({"error":"expected multipart/form-data"})", "application/json");
            return;
        }

        // Достаём поля — точно как у Кости
        std::string chatId, senderId, recipientId, senderName;

        auto chatIt   = req.form.fields.find("chat_id");
        if (chatIt    != req.form.fields.end()) chatId      = chatIt->second.content;

        auto senderIt = req.form.fields.find("sender_id");
        if (senderIt  != req.form.fields.end()) senderId    = senderIt->second.content;

        auto recipIt  = req.form.fields.find("recipient_id");
        if (recipIt   != req.form.fields.end()) recipientId = recipIt->second.content;

        auto nameIt   = req.form.fields.find("sender_name");
        if (nameIt    != req.form.fields.end()) senderName  = nameIt->second.content;

        // Достаём файл
        std::string fileContent;
        std::string originalName;
        std::string contentType;

        auto fileIt = req.form.files.find("file");
        if (fileIt != req.form.files.end()) {
            fileContent  = fileIt->second.content;
            originalName = fileIt->second.filename;
            contentType  = fileIt->second.content_type;
        }

        // Валидация
        if (fileContent.empty() || chatId.empty() || senderId.empty() || recipientId.empty()) {
            res.status = 400;
            res.set_content(R"({"error":"missing required fields: file, chat_id, sender_id, recipient_id"})", "application/json");
            return;
        }

        if (fileContent.size() > MAX_SIZE) {
            res.status = 400;
            res.set_content(R"({"error":"file too large, max 50MB"})", "application/json");
            return;
        }

        // Определяем тип
        MediaInfo info = detectMediaInfo(contentType);

        // Сохраняем: uploads/media/{subdir}/{date}/{uuid}{ext}
        std::string datePath = getDatePath();
        std::string dirPath  = UPLOAD_DIR + "/" + info.subDir + "/" + datePath;
        fs::create_directories(dirPath);

        std::string uuid     = generateUUID();
        std::string filename = uuid + info.ext;
        std::string filePath = dirPath + "/" + filename;

        std::ofstream outFile(filePath, std::ios::binary);
        outFile.write(fileContent.data(), fileContent.size());
        outFile.close();

        std::string mediaUrl = BASE_URL + "/uploads/media/" + info.subDir + "/" + datePath + "/" + filename;
        long long   fileSize = static_cast<long long>(fileContent.size());

        // Стучимся в Go — точно как Костя
        httplib::Client goClient(GO_URL);

        std::string body =
            R"({"chat_id":")"      + chatId +
            R"(","sender_id":")"   + senderId +
            R"(","recipient_id":")"+ recipientId +
            R"(","media_url":")"   + mediaUrl +
            R"(","type":")"        + info.msgType +
            R"(","file_name":")"   + jsonEscape(originalName) +
            R"(","file_size":)"    + std::to_string(fileSize) +
            R"(,"sender_name":")"  + jsonEscape(senderName) +
            R"("})";

        httplib::Headers headers = {
            {"Content-Type", "application/json"},
            {"Authorization", req.get_header_value("Authorization")}
        };

        auto goRes = goClient.Post("/chat/messages/media", headers, body, "application/json");

        if (!goRes || goRes->status != 201) {
            std::string errDetail = goRes ? goRes->body : "no response from go server";
            res.status = 500;
            res.set_content(R"({"error":"failed to create message in db","detail":")" + jsonEscape(errDetail) + R"("})", "application/json");
            return;
        }

        res.set_content(goRes->body, "application/json");
        std::cout << "[media] saved: " << filePath
                  << " type=" << info.msgType
                  << " size=" << fileSize << std::endl;
    });

    std::cout << "Media server running on port 9091" << std::endl;
    svr.listen("0.0.0.0", 9091);
    return 0;
}
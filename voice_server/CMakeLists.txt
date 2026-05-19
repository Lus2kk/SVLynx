cmake_minimum_required(VERSION 3.16)
project(voice_server)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

include_directories(.)
include_directories(./whisper.cpp)
include_directories(./whisper.cpp/include)
include_directories(./whisper.cpp/ggml/include)

add_executable(voice_server voice_server.cpp)

target_link_libraries(voice_server pthread)
target_link_libraries(voice_server ${CMAKE_SOURCE_DIR}/whisper.cpp/build/src/libwhisper.dylib)

if(APPLE)
    target_link_libraries(voice_server "-framework CoreFoundation")
    target_link_libraries(voice_server "-framework Accelerate")
endif()

set_target_properties(voice_server PROPERTIES
    BUILD_RPATH "${CMAKE_SOURCE_DIR}/whisper.cpp/build/src"
    INSTALL_RPATH "${CMAKE_SOURCE_DIR}/whisper.cpp/build/src"
)
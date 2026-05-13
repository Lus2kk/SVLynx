cmake_minimum_required(VERSION 3.16)
project(media_server)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

include_directories(.)

add_executable(media_server media_server.cpp)

target_link_libraries(media_server pthread)

if(APPLE)
    target_link_libraries(media_server "-framework CoreFoundation")
endif()
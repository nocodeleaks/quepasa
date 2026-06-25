# Go Audio Package Code Summary and Re-engagement Prompt

This document provides a comprehensive summary of the Go `audio` package, detailing its structure, functionalities, and external dependencies. It's designed to be used as a prompt for AI assistance to quickly re-engage with the codebase context.

---

## **Prompt for AI Assistance:**

"I am working on a Go project in the `src/audio` package. This package is responsible for audio processing, including information extraction, transcoding, and waveform generation. Below is a detailed summary of its current state, structure, and external dependencies.

**Please review this summary. If I provide a new request, assume this context and propose solutions or modifications based on the described architecture. My goal is to maintain modularity, efficiency, and robustness.**

---

## **Go Audio Package Overview**

**Package Name:** `audio`

**Purpose:** Provides utilities for handling audio data, specifically focusing on obtaining audio information, transcoding formats, and generating visual waveforms.

**Key Features:**
* **External Tool Availability Check:** Statically checks and caches the availability of `ffmpeg` and `ffprobe` executables in the system's PATH.
* **Audio Information Extraction:** Uses `ffprobe` to extract detailed information (duration, channels, sample rate, MIME type) from various audio formats.
* **Audio Transcoding:** Leverages `ffmpeg` to transcode audio data to WAV format, particularly for unsupported formats or specific needs (e.g., Beep library compatibility).
* **Waveform Generation:** Generates a 64-byte waveform from audio data, handling different input formats and transcoding as necessary to work with the `github.com/gopxl/beep` library.
* **Structured Error Handling:** Returns `error` types for failures, providing clear messages.
* **Logging:** Integrates `logrus` for structured logging, allowing for better debugging and operational monitoring.

---

## **Package Structure and Files**

The `audio` package is organized into multiple Go files, each with a specific responsibility, promoting modularity.

1.  **`src/audio/tools.go`**
    * **Core Functionalities:** Contains the primary functions `GetAudioInfoFromBytes`, `transcodeToWAV`, and `GenerateWaveform`.
    * **External Tool Management:**
        * `ffmpegAvailable`, `ffprobeAvailable` (global `bool` flags): Store the cached availability status of FFmpeg and FFprobe.
        * `initError` (global `error`): Stores the first error encountered if `ffmpeg` or `ffprobe` are not found.
        * `ffmpegOnce`, `ffprobeOnce` (`sync.Once`): Ensures that the `exec.LookPath` check for each tool runs only once across multiple calls.
        * `IsFFMpegAvailable() bool`: Public method to check/get the cached availability of FFmpeg.
        * `IsFFProbeAvailable() bool`: Public method to check/get the cached availability of FFprobe.
        * `AreAudioToolsAvailable() bool`: Public convenience method to check if *both* FFmpeg and FFprobe are available.
        * `GetInitError() error`: Public method to retrieve the initial error encountered during tool availability checks.
    * **Dependencies:** Imports `bytes`, `encoding/json`, `fmt`, `io`, `math`, `os`, `os/exec`, `strconv`, `strings`, `sync`, `time`, `github.com/gopxl/beep/v2`, `github.com/gopxl/beep/v2/mp3`, `github.com/gopxl/beep/v2/wav`, and `github.com/sirupsen/logrus`.

2.  **`src/audio/audio_format.go`**
    * **Type Definition:** Defines the `AudioFormat` string type and its associated constants (e.g., `FormatMP3`, `FormatWAV`).
    * **Format Detection:** Includes utility functions `detectAudioFormat([]byte) AudioFormat` for guessing audio format from byte headers.
    * **MIME Type Check:** Provides `IsAudioMIMEType(string) bool` for MIME type validation.
    * **Dependencies:** Imports `bytes` and `strings`.

3.  **`src/audio/ffprobe_result.go`**
    * **Data Structure:** Defines the `FFProbeResult` struct, used for unmarshaling the JSON output from `ffprobe` commands. This struct captures format-level and stream-level details.

4.  **`src/audio/audio_info.go`**
    * **Data Structure:** Defines the `AudioInfo` struct, which encapsulates parsed audio information (duration, channels, sample rate, MIME type) in a cleaner format for application use.
    * **Dependencies:** Imports `time`.

---

## **External Dependencies**

* **`github.com/gopxl/beep`**: A low-level audio library for Go, used for decoding audio streams (MP3, WAV) and resampling for waveform generation.
* **`github.com/sirupsen/logrus`**: A structured logger for Go, used throughout the package for informative, warning, and error logging.
* **`ffmpeg` (System Executable)**: External multimedia framework used for transcoding audio.
* **`ffprobe` (System Executable)**: Companion tool to `ffmpeg`, used for analyzing multimedia streams and extracting metadata.

---

## **Usage Notes**

* The `logentry` variable needs to be properly initialized (e.g., setting formatter, output, level) in the main application or in a dedicated `init()` function within the `media` package itself if a default setup is desired.
* It's recommended to call `media.AreAudioToolsAvailable()` at the start of your application (or before using any functions that depend on FFmpeg/FFprobe) to check prerequisites. If it returns `false`, `media.GetInitError()` can be called to retrieve the detailed reason.
* Temporary files are created and cleaned up for `ffprobe` and `ffmpeg` operations to handle byte-slice inputs.

---
"
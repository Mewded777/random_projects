Go Weather CLI Application

A lightweight, high-performance Command Line Interface (CLI) application built in Go that fetches real-time weather information for any city using the OpenWeatherMap API.

This project demonstrates secure API integrations in Go by utilizing network protocol handling (`https`), URL query encoding, and local system environment isolation to safely manage authentication keys.

 Features

- **Real-Time Data**: Pulls live temperature and weather updates directly from OpenWeatherMap.
- **Robust Input Handling**: Safely processes city names containing spaces (e.g., "New York", "San Francisco") using secure query escaping.
- **Secure Architecture**: Implements safe credentials management via environment variable isolation to prevent private key exposure.
- **Pure Standard Library**: Zero third-party dependencies required to run the final script.

---

Prerequisites

Before running this application, ensure you have:
1. **Go** installed on your computer.
2. A free API account and a 32-character API key generated from the [OpenWeatherMap API Portal](https://openweathermap.org).

---

Setup & Installation

#1. Set Up Your API Key
The application reads your API key from your computer's temporary memory environment to maintain file safety. Open your terminal and run the configuration command below:

**On macOS or Linux:**
```bash
export OPENWEATHER_API_KEY="your_actual_32_character_api_key_here"
```

**On Windows (PowerShell):**
```powershell
$env:OPENWEATHER_API_KEY="your_actual_32_character_api_key_here"
```

### 2. Run the Application
Execute the script using the Go toolchain by passing your target city as a trailing command-line argument:

```bash
go run WeatherAppCLI.go Paris
```

*Note: For cities consisting of multiple words, wrap the input string cleanly inside quotation marks:*
```bash
go run WeatherAppCLI.go "Addis Ababa"
```

---

Expected Terminal Output

```text
Weather in Paris:
Temperature: 18.50°C
Description: scattered clouds
```

---

Development & Repository Safety
This project is configured to enforce security best practices:
- Your core source file contains **no hardcoded authentication keys**.
- A `.gitignore` rule is actively implemented to ensure local `.env` files are permanently blocked from accidental repository synchronization or staging.

#!/usr/bin/env bash
# For Linux/macOS. Downloads & runs the latest release binary.

# --- Configuration ---
OWNER="hawshemi"
REPO="sni-finder"

# --- Detect OS (Linux or macOS) ---
os_type=$(uname -s)
if [ "$os_type" = "Darwin" ]; then
    platform="darwin"
elif [ "$os_type" = "Linux" ]; then
    platform="linux"
else
    echo "Unsupported OS: $os_type. This script only supports Linux and macOS."
    exit 1
fi

# --- Detect CPU Architecture ---
cpu_arch=$(uname -m)
if [ "$cpu_arch" = "x86_64" ]; then
    arch="amd64"
elif [ "$cpu_arch" = "arm64" ] || [ "$cpu_arch" = "aarch64" ]; then
    arch="arm64"
else
    echo "Unsupported architecture: $cpu_arch"
    exit 1
fi

# --- Construct Expected Asset Name ---
asset_name="${REPO}-${platform}-${arch}"
echo "Detected OS: $platform, Architecture: $arch"
echo "Looking for asset: $asset_name"

# --- Fetch Latest Release Info from GitHub ---
API_URL="https://api.github.com/repos/${OWNER}/${REPO}/releases/latest"
release_json=$(curl -s "$API_URL")

# --- Extract the Download URL without jq ---
# This sed/grep pipeline works as follows:
# 1. sed prints lines starting from the one with our asset "name" up to the first occurrence of "browser_download_url"
# 2. grep filters the line containing "browser_download_url"
# 3. sed extracts the URL value.
download_url=$(echo "$release_json" \
  | sed -n '/"name": *"'"$asset_name"'"/,/"browser_download_url"/p' \
  | grep "browser_download_url" \
  | head -n 1 \
  | sed -E 's/.*"browser_download_url": *"([^"]+)".*/\1/')

if [ -z "$download_url" ]; then
    echo "Error: Could not find asset \"$asset_name\" in the latest release."
    exit 1
fi

echo "Downloading $asset_name from:"
echo "$download_url"
curl -L -o "$asset_name" "$download_url"

# --- Set Execute Permissions & Run the Binary ---
chmod +x "$asset_name"
echo "$asset_name Downloaded. Run with: ./"$asset_name --addr ip" "


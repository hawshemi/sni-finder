# SNI Finder

This script will scan all domains with `TLS 1.3` and `h2` enabled on your VPS IP address range. These domains are useful for SNI domain names in various configurations and tests.

When you begin the scan, two files are created: `results.txt` contains the output log, while `domains.txt` only contains the domain names.

It is recommended to run this scanner locally _(with your residential internet)_. It may cause VPS to be flagged if you run a scanner in the cloud.


## Run

### Linux/Mac OS:

1.
    ```
    curl -L "https://raw.githubusercontent.com/hawshemi/sni-finder/main/sni-finder-run.sh" -o sni-finder-run.sh && chmod +x sni-finder-run.sh && bash sni-finder-run.sh
    ```
2. 
    ```
    ./sni-finder -addr ip
    ```

### Windows:

1. Download from [Releases](https://github.com/hawshemi/SNI-Finder/releases/latest).
2. Open `CMD` or `Powershell` in the directory.
3.
    ```
    .\sni-finder-windows-amd64.exe -addr ip
    ```

#### Replace `ip` with your VPS IP Address.


## Build

### Prerequisites

#### Install `wget`:
```
sudo apt install -y wget
```

#### Run this script to install `Go` & other dependencies _(Debian & Ubuntu)_:

    wget "https://raw.githubusercontent.com/hawshemi/tools/main/go-installer/go-installer.sh" -O go-installer.sh && chmod +x go-installer.sh && bash go-installer.sh

- Reboot is recommended.


#### Then:

#### 1. Clone the repository
```
git clone https://github.com/hawshemi/sni-finder.git 
```

#### 2. Navigate into the repository directory
```
cd sni-finder
```

#### 3. Initiate and download deps
```
go mod init sni-finder && go mod tidy
```

#### 4. Build
```
CGO_ENABLED=0 go build
```

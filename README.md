# SNI Finder

This script will scan all domains with `TLS 1.3` and `h2` enabled on your VPS IP address range. These domains are useful for SNI domain names in various configurations and tests.

When you begin the scan, two files are created: `results.txt` contains the output log, while `domains.txt` only contains the domain names.

It is recommended to run this scanner locally _(with your residential internet)_. It may cause VPS to be flagged if you run a scanner in the cloud.


## Run

### Linux/Mac OS:

1.
    ```
    curl -L "https://raw.githubusercontent.com/hawshemi/dns-tester/main/sni-finder-run.sh" -o sni-finder-run.sh && chmod +x sni-finder-run.sh && bash sni-finder-run.sh
    ```
2. 
    ```
    ./SNI-Finder -addr ip
    ```

### Windows:

1. Download from [Releases](https://github.com/hawshemi/SNI-Finder/releases/latest).
2. Open `CMD` or `Powershell` in the directory.
3.
    ```
    .\SNI-Finder-windows-amd64.exe -addr ip
    ```

#### Replace `ip` with your VPS IP Address.


## Build

### Prerequisites

#### Install `wget`:
```
sudo apt install -y wget
```

#### First run this script to install `Go` & other dependencies _(Debian & Ubuntu)_:
```
wget "https://raw.githubusercontent.com/hawshemi/SNI-Finder/main/install-go.sh" -O install-go.sh && chmod +x install-go.sh && bash install-go.sh
```
- Reboot is recommended.


#### Then:

#### 1. Clone the repository
```
git clone https://github.com/hawshemi/SNI-Finder.git 
```

#### 2. Navigate into the repository directory
```
cd SNI-Finder 
```

#### 3. Download and install `logrus` package
```
go get github.com/sirupsen/logrus
```

#### 4. Build
```
CGO_ENABLED=0 go build
```

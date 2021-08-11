# fastcom-exporter

Exports [Fast.com](https://fast.com) metrics in the prometheus format, caching the results.

## Install

**homebrew**:

```sh
brew install caarlos0/tap/fastcom-exporter
```

**docker**:

```sh
docker run --rm -p 9877:9877 caarlos0/fastcom-exporter
```

**apt**:

```sh
echo 'deb [trusted=yes] https://repo.caarlos0.dev/apt/ /' | sudo tee /etc/apt/sources.list.d/caarlos0.list
sudo apt update
sudo apt install fastcom-exporter
```

**yum**:

```sh
echo '[caarlos0]
name=caarlos0
baseurl=https://repo.caarlos0.dev/yum/
enabled=1
gpgcheck=0' | sudo tee /etc/yum.repos.d/caarlos0.repo
sudo yum install fastcom-exporter
```

**deb/rpm/apk**:

Download the `.apk`, `.deb` or `.rpm` from the [releases page][releases] and install with the appropriate commands.

**manually**:

Download the pre-compiled binaries from the [releases page][releases] or clone the repo build from source.

[releases]: https://github.com/caarlos0/fastcom-exporter/releases

## Stargazers over time

[![Stargazers over time](https://starchart.cc/caarlos0/fastcom-exporter.svg)](https://starchart.cc/caarlos0/fastcom-exporter)

git config core.hooksPath .github/hooks

dpkg --add-architecture arm64
apt update
apt install -y gcc-aarch64-linux-gnu libasound2-dev libgl1-mesa-dev xorg-dev libasound2-dev:arm64 libxxf86vm-dev:arm64 libxinerama-dev:arm64 libxi-dev:arm64 libxcursor-dev:arm64 libxrandr-dev:arm64
wget https://github.com/dulli/go-rpi-ws281x/releases/latest/download/rpi_ws281x.gz
tar -xvf rpi_ws281x.gz -C /
rm rpi_ws281x.gz

/usr/local/py-utils/bin/pipx install poetry
/usr/local/py-utils/bin/poetry install


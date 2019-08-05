# dont edit with vs code - complains about windows line endings in linux
version=$(git describe --tags);
docker build -t mgerb/mgdiscord:latest .;
docker tag mgerb/mgdiscord:latest mgerb/mgdiscord:$version;

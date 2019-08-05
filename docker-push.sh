version=$(git describe --tags);
docker push mgerb/mgdiscord:latest;
docker push mgerb/mgdiscord:$version;

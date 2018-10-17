# Install dependencies
sudo yum update
sudo yum install -y docker git tmux
sudo usermod -a -G docker ec2-user

# Install docker compose
sudo curl -L https://github.com/docker/compose/releases/download/1.22.0/docker-compose-`uname -s`-`uname -m` | sudo tee /usr/local/bin/docker-compose > /dev/null
sudo chmod +x /usr/local/bin/docker-compose
sudo service docker start

# Clone repositories
git clone https://github.com/iampigeon/pigeon-mqtt.git
git clone https://github.com/iampigeon/pigeon-http.git
git clone https://github.com/iampigeon/pigeon.git

#!/bin/bash -ex
exec > >(tee /var/log/cloud-init-output.log|logger -t user-data -s 2>/dev/console) 2>&1

# Install AWS CLI first
apt-get update -y
apt-get install -y curl wget unzip jq ca-certificates gnupg lsb-release
curl https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip -o /tmp/awscliv2.zip
unzip -q /tmp/awscliv2.zip -d /tmp
/tmp/aws/install
rm -rf /tmp/aws /tmp/awscliv2.zip

# Get region from instance metadata
TOKEN=$(curl -X PUT "http://169.254.169.254/latest/api/token" -H "X-aws-ec2-metadata-token-ttl-seconds: 60")
REGION=$(curl -s -H "X-aws-ec2-metadata-token: $TOKEN" http://169.254.169.254/latest/meta-data/placement/region | tr -d '\n')
aws configure set region $REGION

# Install dependencies
apt-get update
apt-get install -y curl wget unzip jq ca-certificates gnupg lsb-release

# Install Docker
install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
chmod a+r /etc/apt/keyrings/docker.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
apt-get update
apt-get install -y docker-ce docker-ce-cli containerd.io

# Install CloudWatch agent
wget https://s3.amazonaws.com/amazoncloudwatch-agent/ubuntu/amd64/latest/amazon-cloudwatch-agent.deb
dpkg -i amazon-cloudwatch-agent.deb
rm amazon-cloudwatch-agent.deb

# CloudWatch config
curl -L https://github.com/GhostConductor/ghostconductor/releases/latest/download/cloudwatch-config.json \
  -o /opt/aws/amazon-cloudwatch-agent/bin/config.json
/opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl \
  -a fetch-config -m ec2 -c file:/opt/aws/amazon-cloudwatch-agent/bin/config.json -s

# Create ghostconductor directories
mkdir -p /opt/ghostconductor/bin
mkdir -p /opt/ghostconductor/jobs
mkdir -p /opt/ghostconductor/repos
mkdir -p /opt/ghostconductor/etc/prompts/intent
mkdir -p /opt/ghostconductor/etc/context
mkdir -p /opt/ghostconductor/shared

# Download ghostconductor binary
curl -L https://github.com/GhostConductor/ghostconductor/releases/latest/download/ghostconductor-linux-amd64.tar.gz \
  -o /tmp/ghostconductor.tar.gz
tar -xzf /tmp/ghostconductor.tar.gz -C /opt/ghostconductor/bin
chmod +x /opt/ghostconductor/bin/ghostconductor
rm -f /tmp/ghostconductor.tar.gz

# Download prompts and context
curl -L https://github.com/GhostConductor/ghostconductor/releases/latest/download/prompts.tar.gz \
  -o /tmp/prompts.tar.gz
curl -L https://github.com/GhostConductor/ghostconductor/releases/latest/download/context.tar.gz \
  -o /tmp/context.tar.gz
tar -xzf /tmp/prompts.tar.gz -C /opt/ghostconductor/etc
tar -xzf /tmp/context.tar.gz -C /opt/ghostconductor/etc
touch /opt/ghostconductor/etc/CONTEXT.md
rm -f /tmp/prompts.tar.gz /tmp/context.tar.gz

# Download policy files
curl -L https://github.com/GhostConductor/ghostconductor/releases/latest/download/network-policy.json \
  -o /opt/ghostconductor/etc/network-policy.json
curl -L https://github.com/GhostConductor/ghostconductor/releases/latest/download/container-policy.json \
  -o /opt/ghostconductor/etc/container-policy.json

# Download systemd service files
curl -L https://github.com/GhostConductor/ghostconductor/releases/latest/download/ghostconductor.env \
  -o /etc/systemd/system/ghostconductor.env
curl -L https://github.com/GhostConductor/ghostconductor/releases/latest/download/ghostconductor.service \
  -o /etc/systemd/system/ghostconductor.service

# Set ownership
chown -R ubuntu:ubuntu /opt/ghostconductor

# Enable Docker
systemctl enable docker
systemctl start docker

# Add ubuntu to docker group
usermod -aG docker ubuntu

# Pull ghost image
docker pull ghcr.io/ghostconductor/ghost:latest

# Enable and start ghostconductor
systemctl daemon-reload
systemctl enable ghostconductor
systemctl start ghostconductor

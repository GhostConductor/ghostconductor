# GhostConductor — Standalone Server Deployment

A single EC2 instance running GhostConductor. One server, one project. Ideal for personal use or small teams.

## What It Creates

- EC2 instance running Ubuntu 24.04 (latest, via SSM parameter)
- GhostConductor binary installed and running as a systemd service
- Docker installed and running
- CloudWatch agent configured for logs
- Isolated `ghostconductor` Docker network for agent containers

## Prerequisites

Before deploying you need:

- An AWS account with EC2 permissions
- A VPC with a public subnet
- An EC2 key pair
- An IAM instance profile with the following permissions:
  - `AmazonSSMManagedInstanceCore`
  - `CloudWatchAgentServerPolicy`
- A security group (see recommendations below)

## Security Group Recommendations

GhostConductor runs on port `7777`. **Never open this port to `0.0.0.0/0`.**

Recommended inbound rules:

| Port | Protocol | Source | Purpose |
|------|----------|--------|---------|
| 7777 | TCP | Your IP only | GhostConductor UI |
| 22 | TCP | Your IP only | SSH access |

Recommended outbound rules:

| Port | Protocol | Destination | Purpose |
|------|----------|-------------|---------|
| 443 | TCP | 0.0.0.0/0 | GitHub, Docker Hub, AI provider APIs, package repos |
| 53 | UDP | 0.0.0.0/0 | DNS |

> Agent containers run on an isolated Docker network with ICC disabled. Outbound access is further restricted via `network-policy.json`.

## IAM Instance Profile

Your instance profile needs the following permissions:

- `AmazonSSMManagedInstanceCore` — SSM access
- `CloudWatchAgentServerPolicy` — CloudWatch logs

## Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `InstanceType` | EC2 instance type | `t3a.small` |
| `KeyName` | EC2 key pair name | — |
| `IamInstanceProfile` | IAM instance profile name | — |
| `SubnetId` | Public subnet ID | — |
| `SecurityGroupId` | Security group ID | — |

## Deploy

Download the template and deploy:

```bash
curl -L https://github.com/GhostConductor/ghostconductor/releases/latest/download/server.yaml -o server.yaml

aws cloudformation deploy \
  --template-file server.yaml \
  --stack-name gc-server \
  --parameter-overrides \
    KeyName=your-key-pair \
    IamInstanceProfile=your-instance-profile \
    SubnetId=subnet-xxxxxxxx \
    SecurityGroupId=sg-xxxxxxxx \
  --region us-west-2
```

## Accessing GhostConductor

Once deployed, find your instance's public IP in the CloudFormation outputs:

http://<public-ip>:7777

## Logs

CloudWatch log groups:
- `/ec2/ghostconductor/cloud-init` — userdata bootstrap logs
- `/ec2/ghostconductor/ghostconductor` — application logs

## Configuration

Edit `/etc/systemd/system/ghostconductor.env` to configure:

| Variable | Default | Purpose |
|----------|---------|---------|
| `GC_BASE_PATH` | `/opt/ghostconductor` | Base data directory |
| `GC_PORT` | `7777` | HTTP port |

> API keys are set via the GhostConductor UI — never store them in env files.

After editing, restart the service:

```bash
sudo systemctl restart ghostconductor
```

## Customization

- **Prompts** — edit or add files in `/opt/ghostconductor/etc/prompts/`
- **Context** — edit `/opt/ghostconductor/etc/CONTEXT.md`
- **Network policy** — edit `/opt/ghostconductor/etc/network-policy.json` to control agent outbound access
- **Container policy** — edit `/opt/ghostconductor/etc/container-policy.json` to control resource limits
- **Fork** — fork the [GhostConductor repo](https://github.com/GhostConductor/ghostconductor) to customize and ship your own release

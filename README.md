![Ghost Conductor - AI Agent Orchestration](cmd/ghostconductor/ui/dist/images/gc_hero.png)

**Ghost Conductor** is a container orchestration platform that runs a fleet of autonomous AI agents in disposable, sandboxed containers.

Set your context, choose an intent, and describe the task — ghosts will checkout code and open pull requests for your review.

Support for Anthropic, OpenAI, and Google models. Bring your own API keys and monitor costs per job, per provider, per model.

## Mac

```bash
brew tap GhostConductor/ghostconductor
brew install --cask ghostconductor
ghostconductor
```

## Server (AWS)

Deploy to AWS using the CloudFormation template:

```bash
aws cloudformation deploy \
  --template-file https://github.com/GhostConductor/ghostconductor/releases/latest/download/server.yaml \
  --stack-name gc-server \
  --parameter-overrides \
    KeyName=your-key-pair \
    IamInstanceProfile=your-instance-profile \
    SubnetId=subnet-xxxxxxxx \
    SecurityGroupId=sg-xxxxxxxx \
  --region us-west-2
```

See [server deployment guide](deploy/cf/standalone/server.md) for prerequisites and security group recommendations.

## Customization

Fork this repo to customize prompts, context templates, and policies — then ship your own release.

- **Prompts** — `prompts/intent/` — one `.md` file per intent
- **Context** — `context/` — context templates loaded into every job
- **Network policy** — `config/network-policy.json` — allowed outbound domains for agent containers
- **Container policy** — `config/container-policy.json` — resource limits and security settings

## Agents

- [ghost](https://github.com/GhostConductor/ghost) — the agent runtime image

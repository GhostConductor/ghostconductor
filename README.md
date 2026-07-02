![Ghost Conductor - AI Agent Orchestration](cmd/ghostconductor/ui/dist/images/gc_hero.png)

## What is Ghost Conductor?
Ghost Conductor is an AI agent orchestration platform that runs a fleet of autonomous software engineers in disposable, sandboxed containers.
- Set your context, choose an intent, and describe the task — Ghosts work on your tasks, creating pull requests for review.
- Run multiple Ghosts simultaneously or one Ghost at a time.  When the job is done, the container is destroyed — leaving only the code it wrote.
- Support Anthropic, OpenAI, and Google models — bring your own API keys and mix and match models depending on the task.

## Get Started

### Mac

```bash
brew tap GhostConductor/ghostconductor
brew install --cask ghostconductor
ghostconductor
```

### Server

Deploy to AWS using CloudFormation:

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

See [server deployment guide](deploy/cf/standalone/server.md) for prerequisites and security group recommendations.

## Customization

Fork this repo to customize prompts, context templates, and policies — then ship your own release.

- **Prompts** — `prompts/intent/` — one `.md` file per intent
- **Context** — `context/` — context templates loaded into every job
- **Network policy** — `config/network-policy.json` — allowed outbound domains for agent containers
- **Container policy** — `config/container-policy.json` — resource limits and security settings

## Agents

- [ghost](https://github.com/GhostConductor/ghost) — the agent runtime image

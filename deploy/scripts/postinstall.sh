#!/bin/bash -e

mkdir -p /opt/ghostconductor/etc/prompts/intent
mkdir -p /opt/ghostconductor/etc/context
mkdir -p /opt/ghostconductor/jobs
mkdir -p /opt/ghostconductor/repos
mkdir -p /opt/ghostconductor/shared

tar -xzf /tmp/prompts.tar.gz -C /opt/ghostconductor/etc
tar -xzf /tmp/context.tar.gz -C /opt/ghostconductor/etc

touch /opt/ghostconductor/etc/CONTEXT.md

curl -L https://github.com/GhostConductor/ghostconductor/releases/latest/download/network-policy.json \
  -o /opt/ghostconductor/etc/network-policy.json
curl -L https://github.com/GhostConductor/ghostconductor/releases/latest/download/container-policy.json \
  -o /opt/ghostconductor/etc/container-policy.json

rm -f /tmp/prompts.tar.gz
rm -f /tmp/context.tar.gz

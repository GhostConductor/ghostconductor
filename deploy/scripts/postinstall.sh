#!/bin/bash -e

mkdir -p /opt/ghostconductor/etc
mkdir -p /opt/ghostconductor/jobs
mkdir -p /opt/ghostconductor/repos
mkdir -p /opt/ghostconductor/shared

tar -xzf /tmp/prompts.tar.gz -C /opt/ghostconductor/etc
tar -xzf /tmp/context.tar.gz -C /opt/ghostconductor/etc

touch /opt/ghostconductor/etc/CONTEXT.md

rm -f /tmp/prompts.tar.gz
rm -f /tmp/context.tar.gz

# UI Testing Guide

## Overview

This guide explains how to test the Brain desktop UI against the daemon-backed skills and testing flows.

## What to Verify

- The UI loads data from the daemon.
- Search and filters behave as expected.
- Create, edit, and delete actions round-trip correctly.
- Errors are shown clearly when the daemon is unavailable.
- Sync state matches the daemon and CLI.

## Test Layers

### 1. Smoke Checks

- Open the UI.
- Confirm the main dashboard loads.
- Confirm the skills and testing sections render.

### 2. Interaction Checks

- Search for a known item.
- Filter by type or status.
- Open forms and confirm validation works.

### 3. Sync Checks

- Create or update an item in the UI.
- Confirm the daemon persists the change.
- Confirm the CLI reports the same state.

## Recommended Tools

- Playwright for browser automation.
- Go integration tests for daemon-backed workflows.
- The Brain CLI for cross-checking state.

## Rules

- Do not rely on manual shell scripts for production validation.
- Keep the daemon as the source of truth.
- Prefer automated checks over ad hoc clicks.

## Outcome

A UI test should confirm that the desktop surface, daemon, and CLI all agree on the same state.

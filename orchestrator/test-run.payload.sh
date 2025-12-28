#!/bin/bash

TEST_TARGET_ADDR=127.0.0.1:2222 SSH_USER=pi SSH_KEY_PATH=../scripts/ssh-test/id_ed25519 LOCAL_TESTS_ROOT=../mock-tests REMOTE_DIR=/tmp/test_run TEST_PRODUCT=product_A TEST_PLAN=smoke go run ./cmd/orchestrator

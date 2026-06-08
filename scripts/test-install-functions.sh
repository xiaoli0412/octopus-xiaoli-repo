#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

export OCTOPUS_INSTALL_SH_SOURCE_ONLY=1
# shellcheck source=scripts/install.sh
source "${ROOT_DIR}/scripts/install.sh"

fail_test() {
    printf '[FAIL] %s\n' "$1" >&2
    exit 1
}

assert_eq() {
    local actual="$1"
    local expected="$2"
    local label="$3"

    if [[ "$actual" != "$expected" ]]; then
        fail_test "${label}: got '${actual}', want '${expected}'"
    fi
}

test_disabled_dockerhub_images() {
    is_disabled_dockerhub_image "xiaoli0412/octopus-xiaoli-repo:v1" || fail_test "Docker Hub shorthand image should be disabled"
    is_disabled_dockerhub_image "docker.io/xiaoli0412/octopus-xiaoli-repo:v1" || fail_test "docker.io image should be disabled"
    if is_disabled_dockerhub_image "ghcr.io/xiaoli0412/octopus-xiaoli-repo:v1"; then
        fail_test "GHCR image should be allowed"
    fi
}

test_resolve_data_dir_honors_explicit_env() {
    local called="0"
    docker() {
        called="1"
        return 1
    }

    OCTOPUS_DATA_DIR="/explicit/data"
    assert_eq "$(resolve_data_dir "/explicit/data" "octopus")" "/explicit/data" "explicit data dir"
    assert_eq "$called" "0" "explicit data dir should not inspect docker"
    unset OCTOPUS_DATA_DIR
    unset -f docker
}

test_resolve_data_dir_reuses_existing_mount() {
    docker() {
        if [[ "${1:-}" == "inspect" ]]; then
            printf '/root/octopus-data\n'
            return 0
        fi
        return 1
    }

    assert_eq "$(resolve_data_dir "./data" "octopus")" "/root/octopus-data" "existing container data dir"
    unset -f docker
}

test_resolve_external_port_reuses_owned_port() {
    port_in_use() {
        [[ "$1" == "1088" ]]
    }
    port_owned_by_container() {
        [[ "$1" == "1088" && "$2" == "octopus" ]]
    }

    assert_eq "$(resolve_external_port "1088" "octopus")" "1088" "owned port reuse"
    unset -f port_in_use
    unset -f port_owned_by_container
}

test_disabled_dockerhub_images
test_resolve_data_dir_honors_explicit_env
test_resolve_data_dir_reuses_existing_mount
test_resolve_external_port_reuses_owned_port

printf '[PASS] install script function tests\n'

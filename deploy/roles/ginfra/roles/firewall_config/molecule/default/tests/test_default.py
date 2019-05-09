import os

import testinfra.utils.ansible_runner

testinfra_hosts = testinfra.utils.ansible_runner.AnsibleRunner(
    os.environ['MOLECULE_INVENTORY_FILE']).get_hosts('all')


def test_hosts_file(host):
    cmd = host.run("firewall-cmd --list-all")
    for line in cmd.stdout.split("\n"):
        if "services" in line:
            assert "glusterfs" in line
    for line in cmd.stdout.split("\n"):
        if " ports:" in line:
            assert "2049/tcp" in line
            assert "54321/tcp" in line
            assert "5900/tcp" in line
            assert "5900-6923/tcp" in line
            assert "5666/tcp" in line
            assert "16514/tcp" in line

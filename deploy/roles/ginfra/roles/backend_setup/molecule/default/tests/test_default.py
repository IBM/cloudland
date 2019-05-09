import os

import testinfra.utils.ansible_runner

testinfra_hosts = testinfra.utils.ansible_runner.AnsibleRunner(
    os.environ['MOLECULE_INVENTORY_FILE']).get_hosts('all')


def test_hosts_file(host):
    cmd = host.run("vgdisplay")
    for line in cmd.stdout.split(" "):
        if "VG Name" in line:
            assert "vg_vdb" in line
            assert "vg_vdc" in line
    cmd = host.run("lvdisplay")
    for line in cmd.stdout.split(" "):
        if "LV Path" in line:
            assert "/dev/vg_vdc/vg_vdc_thinlv1" in line
            assert "/dev/vg_vdc/vg_vdc_thinlv2" in line
            assert "/dev/vg_vdb/thicklv_1" in line
        if "LV Name" in line:
            assert "foo_thinpool" in line
            assert "vg_vdc_thinlv1" in line
            assert "vg_vdc_thinlv2" in line
            assert "thicklv_1" in line

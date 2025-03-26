from netfilterqueue import NetfilterQueue
from scapy.all import *
import os

gateway_mac = None
external_vlan = None

def process_packet(packet):
    scapy_packet = IP(packet.get_payload())

    #print(f"Source IP: {scapy_packet.src}, Destination IP: {scapy_packet.dst}")
    sendp(Ether(dst=gateway_mac)/scapy_packet, iface=external_vlan, verbose=0)

    packet.drop()

if __name__ == "__main__":
    if len(sys.argv) != 4:
        print("Usage: forward_pkt.py <queue_id> <ext_vlan> <gateway_mac>")
        sys.exit(1)

    queue_id = sys.argv[1]
    external_vlan = sys.argv[2]
    gateway_mac = sys.argv[3]

    nfqueue = NetfilterQueue()
    nfqueue.bind(int(queue_id), process_packet)

    try:
        print("Starting NFQueue...")
        nfqueue.run()
    except KeyboardInterrupt:
        print("Stopping NFQueue...")
    finally:
        nfqueue.unbind()

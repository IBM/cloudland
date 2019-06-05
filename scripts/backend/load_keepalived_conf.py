#!/usr/bin/python
"""
forked and modified 
https://github.com/khosrow/lvsm/blob/master/lvsm/modules/kaparser.py
Parse a keepalived configuration file. The config specs are defined in keepalived.conf(5)
"""
from pyparsing import *
import os
import logging
import json

logging.basicConfig(format='[%(levelname)s]: %(message)s')
logger = logging.getLogger('keepalived')

def tokenize_config(configfile):

    LBRACE, RBRACE = map(Suppress, "{}")

    # generic value types
    integer = Word(nums)
    string = Word(printables)

    ip4_address = Regex(r"\d{1,3}(\.\d{1,3}){3}")
    ip6_address = Regex(r"[0-9a-fA-F:]+")
    ip_address = ip4_address | ip6_address

    ip4_class = Regex(r"\d{1,3}(\.\d{1,3}){3}\/\d{1,3}")
    ip6_class = Regex(r"[0-9a-fA-F:]+\/\d{1,3}")
    ip_class = ip4_class | ip6_class
    ip_classaddr = ip_class | ip_address

    scope = oneOf("site link host nowhere global")

    # Parameters to "ip addr add" command
    ipaddr_cmd = (ip_classaddr + Optional("dev" + Word(printables)) +
                  Optional("scope" + scope) + Optional("label" + string)) 

    # Params for ip routes
    ip_route = ("src" + ip_address + Optional("to") + ip_classaddr +
                oneOf("via gw") + ip_address + "dev" + Word(alphanums) +
                "scope" + scope + "table" + Word(alphanums))

    string_or_quoted = (string | quotedString)

    # vrrp_sync_params
    notify_master = ("notify_master" + string_or_quoted)
    notify_backup = ("notify_backup" + string_or_quoted)
    notify_fault = ("notify_fault" + string_or_quoted)
    notify = ("notify" + string_or_quoted)
    smtp_alert = ("smtp_alert")

    # vrrp_instance_params
    use_vmac = ("use_vmac")
    native_ipv6 = ("native_ipv6")
    state = Dict(Group("state" + oneOf("MASTER BACKUP SLAVE")))
    interface = Dict(Group("interface" + Word(printables)))
    track_interface = Dict(Group("track_interface" + LBRACE + OneOrMore(string, stopOn=RBRACE) + RBRACE))
    track_script = Dict(Group("track_script" + LBRACE + OneOrMore(Word(alphanums)) + RBRACE))
    dont_track_primary = ("dont_track_primary")
    mcast_src_ip = Dict(Group("mcast_src_ip" + ip_address))
    unicast_peer = Dict(Group("unicast_peer" + LBRACE + OneOrMore(ip_address) + RBRACE))
    lvs_sync_daemon = Dict(Group("lvs_sync_daemon_interface" + Word(alphanums)))
    garp_master_delay = Dict(Group("garp_master_delay" + integer))
    virtual_router_id = Dict(Group("virtual_router_id" + integer))
    priority = Dict(Group("priority" + integer))
    advert_int = Dict(Group("advert_int" + integer))
    authentication = Dict(Group("authentication" +
                           LBRACE + 
                           "auth_type" + oneOf("PASS AH") +
                           "auth_pass" + string +
                           RBRACE))
    nopreempt = ("nopreempt")
    preempt_delay = Dict(Group("preempt_delay" + integer))
    debug = ("debug")
    virtual_ipaddress = Dict(Group("virtual_ipaddress" +
                              LBRACE +
                              OneOrMore(ipaddr_cmd, stopOn=RBRACE) +
                              RBRACE))
    virtual_ipaddress_excluded = Dict(Group("virtual_ipaddress_excluded" +
                                       LBRACE +
                                       OneOrMore(ipaddr_cmd, stopOn=RBRACE) +
                                       RBRACE))
    virtual_routes = Dict(Group("virtual_routes" +
                           LBRACE +
                           OneOrMore(ip_route) +
                           RBRACE))

    vrrp_instance_params = (state | interface |
                            track_interface | track_script |
                            use_vmac | native_ipv6 | dont_track_primary |
                            nopreempt | debug | preempt_delay |
                            mcast_src_ip | unicast_peer  | lvs_sync_daemon |
                            garp_master_delay | virtual_router_id | priority |
                            virtual_ipaddress | virtual_ipaddress_excluded |
                            virtual_routes | advert_int | authentication |
                            notify_master | notify_backup | notify_fault |
                            notify | smtp_alert)

    vrrp_instance = Dict(Group("vrrp_instance" + string +
                          LBRACE +
                          OneOrMore(vrrp_instance_params) +
                          RBRACE))

    comment = oneOf("# !") + restOfLine
    allconfig = OneOrMore(vrrp_instance)
    allconfig.ignore(comment)

    try: 
#        print(configfile)
        tokens = allconfig.parseString(configfile)
        print(tokens)
        vni = tokens['vrrp_instance']['interface'].split('-', 1)[1]
        print(vni)
        router = tokens['vrrp_instance'][0]
        print(router)
        os.system("ip netns add %s" % (router))
        if tokens['vrrp_instance']['state'] == 'MASTER':
            os.system("/opt/cloudland/scripts/backend/set_gateway.sh %s %s %s hard" % (router, "169.254.169.250/24", vni))
        elif tokens['vrrp_instance']['state'] == 'SLAVE':
            os.system("/opt/cloudland/scripts/backend/set_gateway.sh %s %s %s hard" % (router, "169.254.169.251/24", vni))
        for i in range(len(tokens['vrrp_instance']['virtual_ipaddress'])):
            if i % 3 == 2:
                device = tokens['vrrp_instance']['virtual_ipaddress'][i].split('-', 1)
                print(device)
                if device[0] == 'te':
                    os.system("/opt/cloudland/scripts/backend/create_veth.sh %s ext-%s te-%s" % (router, device[1], device[1]))
                    print("/opt/cloudland/scripts/backend/create_veth.sh %s ext-%s te-%s" % (router, device[1], device[1]))
                elif device[0] == 'ti':
                    os.system("/opt/cloudland/scripts/backend/create_veth.sh %s int-%s ti-%s" % (router, device[1], device[1]))
                    print("/opt/cloudland/scripts/backend/create_veth.sh %s int-%s ti-%s" % (router, device[1], device[1]))
                elif device[0] == 'ns':
                    os.system("/opt/cloudland/scripts/backend/create_veth.sh %s ln-%s ns-%s" % (router, device[1], device[1]))
                    print("/opt/cloudland/scripts/backend/create_veth.sh %s ln-%s ns-%s" % (router, device[1], device[1]))
        return tokens
    except ParseException as e:
        logger.error("Exception")
        logger.error(e)
    except ParseFatalException as e:
        logger.error("FatalException")
        logger.error(e)
    else:
#        json_string = json.dumps(tokens.asDict())
#        print(json_string)
        return tokens

def main():
    import sys
    import argparse

    parser = argparse.ArgumentParser(description=__doc__,
                                      usage="%(prog)s [options] filename")

    parser.add_argument("-q", "--quiet",
                        help="Quiet mode. Return 0 on success, 1 on failure.",
                        action="store_true")
    parser.add_argument("-v", "--verbose",
                        help="Verbose mode. Print all tokens on success.",
                        action="store_true")
    parser.add_argument("file", type=argparse.FileType('r'))

    try:
        args = parser.parse_args()
    except IOError as e:
        print(e)
        sys.exit(2)

    try:
        conf = "".join(args.file.readlines())
    except IOError as e:
        print("%s" % e)
        sys.exit(1)

    t = tokenize_config(conf)

    if t:
        if args.verbose:
            print(t.dump())
            print("---------")
        if not args.quiet:
            print("%s parsed OK!" % args.file.name)
        sys.exit(0)        
    else:
        if not args.quiet:
            print("%s didn't parse OK!" % args.file.name)
        sys.exit(1)

if __name__ == "__main__":
    main()

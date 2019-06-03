"""
Parse a keepalived configuration file. The config specs are defined in keepalived.conf(5)
"""
from pyparsing import *
import logging
import parseactions
import json

logging.basicConfig(format='[%(levelname)s]: %(message)s')
logger = logging.getLogger('keepalived')

def tokenize_config(configfile):

    LBRACE, RBRACE = map(Suppress, "{}")

    # generic value types
    email_addr = Regex(r"[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,4}")
    integer = Word(nums)
    string = Word(printables)

    ip4_address = Regex(r"\d{1,3}(\.\d{1,3}){3}")
    ip6_address = Regex(r"[0-9a-fA-F:]+")
    ip_address = ip4_address | ip6_address

    ip4_class = Regex(r"\d{1,3}(\.\d{1,3}){3}\/\d{1,3}")
    ip6_class = Regex(r"[0-9a-fA-F:]+\/\d{1,3}")
    ip_class = ip4_class | ip6_class
    ip_classaddr = ip_address | ip_class

    ip4_range = Regex(r"\d{1,3}(\.\d{1,3}){3}-\d{1,3}")
    # ip6_range = Regex(r"[0-9a-fA-F:]+\/\d{1,3}-\d{1,3}")
    # ip_range = ip6_range | ip4_range

    scope = oneOf("site link host nowhere global")

    # global params
    notification_emails = Dict(Group("notification_email" +
                                     LBRACE +
                                     OneOrMore(email_addr) +
                                     RBRACE))
    notification_email_from = Dict(Group("notification_email_from" + email_addr))
    smtp_server =  Dict(Group("smtp_server" + ip_address))
    smtp_connect_timeout = Dict(Group("smtp_connect_timeout" + integer))
    router_id = Dict(Group("router_id" + string))

    global_params = (notification_emails | notification_email_from | 
                     smtp_server | smtp_connect_timeout | router_id)

    # Parameters to "ip addr add" command
    ipaddr_cmd = (ip_classaddr + Optional("dev" + Word(alphanums)) +
                  Optional("scope" + scope) + Optional("label" + string)) 

    # Params for ip routes
    ip_route = ("src" + ip_address + Optional("to") + ip_classaddr +
                oneOf("via gw") + ip_address + "dev" + Word(alphanums) +
                "scope" + scope + "table" + Word(alphanums))
    black_hole = ("black_hole" + ip_classaddr)

    route_params = (ip_route | black_hole)

    string_or_quoted = (printables | quotedString)


    # vrrp_script params
    vrrp_scr_name = ("script" + quotedString)
    vrrp_scr_interval = ("interval" + integer)
    vrrp_scr_weight = ("weight" + integer)
    vrrp_scr_fall = ("fall" + integer)
    vrrp_scr_rise = ("rise" + integer)

    vrrp_scr_params = (vrrp_scr_name | vrrp_scr_interval | vrrp_scr_weight |
                      vrrp_scr_fall | vrrp_scr_rise)

    # vrrp_sync_params
    sync_group = ("group" + LBRACE + OneOrMore(Word(printables)) + RBRACE)
    notify_master = ("notify_master" + string_or_quoted)
    notify_backup = ("notify_backup" + string_or_quoted)
    notify_fault = ("notify_fault" + string_or_quoted)
    notify = ("notify" + string_or_quoted)
    smtp_alert = ("smtp_alert")

    vrrp_sync_params = (sync_group | notify_fault | notify_backup |
                        notify_master | notify | smtp_alert)

    # vrrp_instance_params
    use_vmac = ("use_vmac")
    native_ipv6 = ("native_ipv6")
    state = Dict(Group("state" + oneOf("MASTER BACKUP")))
    interface = Dict(Group("interface" + Word(alphanums)))
    track_interface = Dict(Group("track_interface" + LBRACE + OneOrMore(Word(alphanums)) + RBRACE))
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
                              OneOrMore(ipaddr_cmd) +
                              RBRACE))
    virtual_ipaddress_excluded = Dict(Group("virtual_ipaddress_excluded" +
                                       LBRACE +
                                       OneOrMore(ipaddr_cmd) +
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

    # virtual_server_group params
    vip_vport = (ip_address + integer)
    vip_range_vport = (ip4_range + integer)
    fwmark = ("fwmark" + integer)

    vserver_group_params = (vip_vport | vip_range_vport | fwmark)

    # virtual_server params
    vserver_group = ("group" + string)
    vserver_vip_vport = (ip_address + integer)
    vserver_fwm = ("fwmark" + integer)
    vserver_id = (vserver_group | vserver_vip_vport | vserver_fwm)

    delay_loop = Dict(Group("delay_loop" + integer))
    lb_algo = ("lb_algo" + oneOf("rr wrr lc wlc lblc sh dh"))
    lb_kind = ("lb_kind" + oneOf("NAT DR TUN"))
    persistence_timeout = ("persistence_timeout" + integer)
    persistence_granularity = ("persistence_granularity" + string) # TODO: this should be a netmask
    protocol = ("protocol" + oneOf("TCP UDP FWM"))
    ha_suspend = ("ha_suspend")
    virtual_host = ("virtual_host" + string)
    alpha = ("alpha")
    omega = ("omega")
    quorom = ("quorom" + integer)
    hysteresis = ("hysteresis" + integer)
    quorom_up = ("quorom_up" + (string | quotedString))
    quorom_down = ("quorom_down" + (string | quotedString))
    sorry_server = ("sorry_server" + ip_address + integer)

    vserver_params = (delay_loop | persistence_timeout | ha_suspend |
                      persistence_granularity | virtual_host | alpha |
                      omega | quorom | hysteresis | quorom_up | quorom_down |
                      sorry_server)

    # real_server_params section
    weight = ("weight" + integer)
    inhibit_on_failure = ("inhibit_on_failure")
    notify_up = ("notify_up" + (string | quotedString))
    notify_down = ("notify_down" + (string | quotedString))

    # http_check_params
    path = ("path" + string)
    digest = ("digest" + Word(hexnums))
    status_code = ("status_code" + integer)

    url_params = (path | status_code | digest)

    url_check = Group("url" +
                      LBRACE +
                      OneOrMore(url_params) +
                      RBRACE)

    connect_port = ("connect_port" + integer)
    bindto = ("bindto" + ip_address)
    connect_timeout = ("connect_timeout" + integer)
    nb_get_retry  = ("nb_get_retry" + integer)
    delay_before_retry = ("delay_before_retry" + integer)

    http_check_params = (connect_port | connect_timeout | bindto |
                         nb_get_retry | delay_before_retry)
    
    http_check = Group(oneOf("HTTP_GET SSL_GET") +
                        LBRACE +
                        OneOrMore(url_check) &
                        ZeroOrMore(http_check_params) &
                        RBRACE)

    tcp_check = Group("TCP_CHECK" +
                      LBRACE +
                      OneOrMore(connect_port | connect_timeout | bindto) +
                      RBRACE)

    # smtp_check params
    connect_ip = ("connect_ip" + ip_address)
    smtp_host = Group("host" +
                      LBRACE +
                      connect_ip +
                      connect_port +
                      bindto +
                      RBRACE)
    retry = Dict(Group("retry" + integer))
    helo_name = Dict(Group("helo_name" + string_or_quoted))
    smtp_check_params = (connect_timeout | retry | delay_before_retry | helo_name )

    smtp_check = Dict(Group("SMTP_CHECK" +
                       LBRACE +
                       Optional(smtp_host) +
                       ZeroOrMore(smtp_check_params) +
                       RBRACE))

    # misc_check params
    misc_path = Dict(Group("misc_path" + string_or_quoted))
    misc_timeout = Dict(Group("misc_timeout" + integer))
    misc_dynamic = ("misc_dynamic")

    misc_check = Dict(Group("MISC_CHECK" +
                       LBRACE +
                       misc_path + 
                       ZeroOrMore(misc_timeout | misc_dynamic) +
                       RBRACE))

    check_type = (http_check | tcp_check | smtp_check | misc_check)

    real_server_params = (weight | inhibit_on_failure | notify_up | notify_down | check_type)
    # Real server block
    real_server = Dict(Group("real_server" + ip_address + integer +
                        LBRACE +
                        OneOrMore(real_server_params) +
                        RBRACE
                        ))


    # Major blocks in the keepalived config

    global_defs = Dict(Group("global_defs" +
                        LBRACE +
                        OneOrMore(global_params) +
                        RBRACE))
    static_ipaddress = Dict(Group("static_ipaddress" +
                             LBRACE +
                             OneOrMore(ipaddr_cmd) +
                             RBRACE))
    static_routes = Dict(Group("static_routes" +
                          LBRACE +
                          OneOrMore(route_params) +
                          RBRACE))
    vrrp_script = Dict(Group("vrrp_script" +
                        LBRACE +
                        OneOrMore(vrrp_scr_params) +
                        RBRACE))

    vrrp_sync_group = Dict(Group("vrrp_sync_group" + string +
                            LBRACE +
                            OneOrMore(vrrp_sync_params) +
                            RBRACE))

    vrrp_instance = Dict(Group("vrrp_instance" + string +
                          LBRACE +
                          OneOrMore(vrrp_instance_params) +
                          RBRACE))

    virtual_server_group = Dict(Group("virtual_server_group" + string +
                                 LBRACE +
                                 OneOrMore(vserver_group_params) + 
                                 RBRACE))

    # virtual_server = Dict(Group("virtual_server" + vserver_id +
    #                        LBRACE &
    #                        # Each([lb_algo, lb_kind, protocol, ZeroOrMore(vserver_params), OneOrMore(real_server) ]) +
    #                        lb_algo &
    #                        lb_kind &
    #                        protocol &
    #                        ZeroOrMore(vserver_params) &
    #                        OneOrMore(real_server) &
    #                        RBRACE
    #                        ))

    virtual_server = Dict(Group("virtual_server" + 
                            ip_address("vip_address") + 
                            integer("vport") + LBRACE & 
                            lb_algo &
                            lb_kind &
                            protocol &
                            ZeroOrMore(vserver_params) &
                            OneOrMore(real_server) &
                            RBRACE
                            ))

    comment = oneOf("# !") + restOfLine

    allconfig = (global_defs &                
                ZeroOrMore(static_ipaddress | static_routes | vrrp_instance |                  
                           vrrp_script | vrrp_sync_group |
                           virtual_server_group | virtual_server) 
                )
    allconfig.ignore(comment)

    try: 
        tokens = allconfig.parseString(configfile)
        print(tokens)
    except ParseException as e:
        logger.error("Exception")
        logger.error(e)
    except ParseFatalException as e:
        logger.error("FatalException")
        logger.error(e)
    else:
        json_string = json.dumps(tokens.asDict())
        print(json_string)
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

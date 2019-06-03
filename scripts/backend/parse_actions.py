"""
forked from
https://github.com/khosrow/lvsm/blob/master/lvsm/modules/parseactions.py
"""

import socket
from pyparsing import *


def validate_ip4(s, loc, tokens):
    """Helper function to validate IPv4 addresses"""
    try:
        socket.inet_pton(socket.AF_INET, tokens[0])
    except socket.error:
        errmsg = "invalid IPv4 address."
        raise ParseFatalException(s, loc, errmsg)

    return tokens

def validate_ip6(s, loc, tokens):
    """Helper function to validate IPv6 adddresses"""
    try:
        socket.ineet_pton(socket.AF_INET6, tokens[0])
    except socket.error:
        errmsg = "invalid IPv6 address."
        raise ParseFatalException(s, loc, errmsg)

    return tokens

def validate_port(s, loc, tokens):
    """Helper function that verifies we have a valid port number"""
    # port = tokens[1]
    port = tokens[0]
    if int(port) < 65535 and int(port) > 0:
        return tokens
    else:
        errmsg = "Invalid port number!"
        raise ParseFatalException(s, loc, errmsg)
    
def validate_scheduler(s, loc, tokens):
    schedulers = ['rr', 'wrr', 'lc', 'wlc', 'lblc', 'lblcr', 'dh', 'sh', 'sed', 'nq']

    if tokens[0][1] in schedulers:
        return tokens
    else:
        errmsg = "Invalid scheduler type!"
        raise ParseFatalException(s, loc, errmsg)

def validate_checktype(s, loc, tokens):
    checktypes = ["connect", "negotiate", "ping", "off", "on", "external", "external-perl"]
    if ((tokens[0][1] in checktypes) or (tokens[0][1].isdigit() and int(tokens[0][1]) > 0)):
        return tokens
    else:
        errmsg = "Invalid checktype!"
        raise ParseFatalException(s, loc, errmsg)

def validate_int(s, loc, tokens):
    """Validate that the token is an integer"""
    try:
        int(tokens[0])
    except ValueError:
        errmsg = "Value must be an integer!"
        raise ParseFatalException(s, loc, errmsg)

def validate_protocol(s, loc, tokens):
    protocols = ['fwm', 'udp', 'tcp']
    if tokens[0][1] in protocols:
        return tokens
    else:
        errmsg = "Invalid protocol!"
        raise ParseFatalException(s, loc, errmsg)

def validate_service(s, loc, tokens):
    services = ["dns", "ftp", "http", "https", "http_proxy", "imap", "imaps"
                ,"ldap", "nntp", "mysql", "none", "oracle", "pop" , "pops"
                , "radius", "pgsql" , "sip" , "smtp", "submission", "simpletcp"]
    if tokens[0][1] in services:
        return tokens
    else:
        errmsg = "Invalid service type!"
        raise ParseFatalException(s, loc, errmsg)

def validate_yesno(s, loc, tokens):
    # if tokens[0] == "yes" or tokens[0] == "no":
    if tokens[0] in ['yes', 'no']:
        return tokens
    else:
        errmsg = "Value must be 'yes' or 'no'"
        raise ParseFatalException(s, loc, errmsg)

def validate_httpmethod(s, loc, tokens):
    if tokens[0][1] in ['GET', 'HEAD']:
        return tokens
    else:
        errmsg = "Value must be 'GET' or 'HEAD'"
        raise ParseFatalException(s, loc, errmsg)

def validate_lbmethod(s, loc, tokens):
    """Validate the load balancing method used for real servers"""
    methods = ['gate', 'masq', 'ipip']
    if tokens[0] in methods:
        return tokens
    else:
        errmsg = "Loadbalancing method must be one of %s " % ', '.join(methods)
        raise ParseFatalException(s, loc, errmsg)

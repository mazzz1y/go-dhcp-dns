# go-dhcp-dns

A DNS proxy server that enables selective DNS resolution through DHCP-provided DNS servers, particularly useful when working with VPNs and captive portals.

## Problem

In some environments, it's challenging to configure selective DNS resolution using DHCP-propagated DNS servers for specific domains. This becomes problematic when:

- Using a local custom DNS server (e.g., through VPN)
- Needing to resolve specific domains through DHCP-provided DNS servers
- Dealing with captive portals that require local DNS resolution
- Managing split-DNS configurations

## Solution

go-dhcp-dns provides a lightweight DNS proxy that:

1. Discovers DNS servers from DHCP using system tools without interfering with system DHCP client
2. Proxies DNS queries to these DHCP-provided servers
3. Implements proper DNS caching with TTL support
4. Allows selective domain resolution through DHCP DNS servers while maintaining VPN connectivity

Getting DNS servers was done in a "hacky" way:

* In the Linux application by looking at `/run/NetworkManager/no-stub-resolv.conf`, which is NetworkManager's resolv.conf path for "real" network resolver.
* In the MacOS application by parsing the ipconfig output . It's a dirty hack, but it works quickly and efficiently, so be careful.

```mermaid
graph TD
    Device[Device]

    subgraph dns-resolution[DNS Resolution]
        local-dns[Local Resolver<br/>127.0.0.1:53]
        go-dhcp-dns[go-dhcp-dns<br/>127.0.0.1:53533]
    end

    subgraph network
        vpn[VPN DNS Servers]
        dhcp[DHCP DNS Servers]
    end

    Device -->|All DNS queries| local-dns
    local-dns -->|Most domains| vpn
    local-dns -->|"captive.apple.com,<br/>portal.example.com"| go-dhcp-dns
    go-dhcp-dns --> dhcp
```
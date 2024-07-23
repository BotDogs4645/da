# Driver Authentication & FMS (LITE)

*name pending*

The goal of this project is to create helper software to:

    - Grant each Romi a 10.0.100.xxx IP for communication
    - Perform FMS-style actions on a Romi robot:
        - Enable/disable
        - Network stats
    - Generate matches
    - Handle scoring ingress

The idea is to use a managed switch to assign VLANs to each Romi to prevent interference. This project is designed around a Cisco Catalyst 3500 series switch, but any managed switch will work. The Catalyst switch was chosen because you can pick one up for under $50 with PoE.

## For Open source

We hope to merge our updated frontend with upstream once it is finished. 
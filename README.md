cjdcmd
======

Cjdcmd is a command line tool for interfacing with [cjdns](https://github.com/cjdelisle/cjdns), a mesh network routing engine designed for security, scalability, speed, and ease of use. Its intent is to allow easy debugging of node and network problems as well as make it easier to work with the cjdns program itself.

Cjdcmd is licensed under the GPL version 3 license, the full text of which is
available in `GPLv3.md`.

What's New
----------

#### Version 0.5.1:

* Added support for .cjdnsadmin file. This file contains details on how to connect to a running cjdns instance, as well as your preferred default configuration file. 
* `cjdnsadmin` will generate a .cjdnsadmin file based on the cjdroute.conf file given to it in the --file flag. If no file is given, it will try using the one specified in a preexisting ~/.cjdnsadmin

#### Version 0.5: 

* Updated to work with the latest version of cjdns that uses UDP to communicate via the admin port.
* `memory` returns the number of bytes of memory allocated by the router.

#### Version 0.4:

* `addpeer` accepts a set of JSON peering details surrounded by single quotes (' ') and will walk you through adding them to your config, along with any additional information you would like to save with it. You can specify which file to read and which file to save to using the -file and -outfile flags, both of which are optional. 
* `addpass` optionally accepts a password, or will generate one if none was supplied, and saves it to your config along with any additional information you may wish to add. You can specify which file to read and which file to save to using the -file and -outfile flags, both of which are optional. 
* `cleanconfig` will read your configuration file, strip the comments, and save it back nicely formatted. You can specify which file to read and which file to save to using the -file and -outfile flags, both of which are optional. 


#### Version 0.3:

* Access to the cjdns configuration file and the cjdns admin port are no longer required by default. These will only be accessed when needed
* The `passgen` command has been added to make creating passwords for new peers easier.
* Hostnames are now resolved wherever possible; this makes the output much more useful.
* When resolving hostnames, cjdcmd first tries your regular system DNS setup, including /etc/hosts, and then it will try HypeDNS if the former failed. I recommend setting hostnames in /etc/hosts as HypeDNS can be unreliable and Internet DNS servers cannot resolve cjdns IPv6 addresses to hostnames.
* Traceroute now supports passing of a cjdns path to it so that you can debug just one particular route. 
* The output of the `ping` command now matches the standard ping programs output more closely.
* The `ping` command will no longer quit on the first error.
* Added -interval flag to the `ping` command so you can set longer delays between pings.
* Added `ip` command to convert a cjdns public key to the corresponding cjdns IPv6 address; great for cleaning up your config file.


Installation
------------

cjdcmd is written in Go. If you already have Go installed, skip the next section.

### Install Go

To install go, check out the [install instructions](http://golang.org/doc/install) or, if you're lucky, you can do a quick install with the information below:

#### Ubuntu (and other Debian variants)

Ubuntu and other Debian variants do not have go 1.1 in their repositories. The easiest way to get it is to use `godeb`. Follow the instructions [here](http://blog.labix.org/2013/06/15/in-flight-deb-packages-of-go), or:

    # 64 bit:
    wget https://godeb.s3.amazonaws.com/godeb-amd64.tar.gz
    # or 32 bit:
    wget https://godeb.s3.amazonaws.com/godeb-386.tar.gz
    
    # untar it:
    tar xzf godeb-*.tar.gz
    
    # execute it:
    ./godeb install
    
    # Clean up:
    rm godeb-*.tar.gz godeb

#### Mac OSX

If you have [Homebrew](http://mxcl.github.com/homebrew/) installed:

    brew install go

If your system isn't listed above, perhaps you can [download pre-compiled binaries](http://code.google.com/p/go/downloads), otherwise you will have to [install from source](http://golang.org/doc/install/source).

#### Configure your folders (optional)

Next you should set up a special directory for all your go sourcecode and compiled programs. This is optional, however I recommend it because this will prevent you from having to use `sudo` every time you need to install an updated package or program. I will be giving a shortened version of the information found on [the official site](http://golang.org/doc/code.html#tmp_2)

First, make the folder where you want everything to be stored. I use the /home/inhies/projects/go but you may use whatever you like:

    $ mkdir -p $HOME/projects/go 

Next we need to tell our system where to look for Go packages and compiled programs. If you changed the folder name in the previous example, then make sure you change them here. You will want to add this to your `~/.profile` file or similar. On Ubuntu 12.10, I had to add it to my `~/.bashrc`:

	export GOPATH=$HOME/projects/go
	export PATH=$PATH:$HOME/projects/go/bin

Now, to make the changes take effect immediately, you can either run `$ source ~/.bashrc` or just paste those two lines you added to the file on your command line. You should now be setup! Try typing `$ echo $GOPATH` and make sure you see the folder you specified. If you have any problems, try re-reading the [official documentation](http://golang.org/doc/code.html#tmp_2).

### Install cjdcmd

Once you have Go installed, installing new programs and packages couldn't be easier. Simply run the following command to have cjdcmd download, build, and install:

    go get github.com/inhies/cjdcmd
	
If f you see no output from that command then everything worked with no errors. To verify that it was successful, run `cjdcmd` and see if it displays some information about the program. If it does you are done! cjdcmd has been downloaded, compiled, and installed. You amy now use it by typing `cjdcmd`.
	
**NOTE:** You may have to be root (use `sudo`) to install Go and cjdcmd.

Updating cjdcmd
---------------

To update your install of cjdcmd, simply run `go get -u github.com/inhies/cjdcmd` and it will automatically update, build, and install it. Just like when you initially installed cjdcmd, if you see no output from that command then everything worked with no errors.
	
Using cjdcmd
------------

Once you have cjdcmd installed you can run it without any arguments to get a list of commands or run it with the flag `--help` to get a list of all support options. Currently cjdcmd offers the following commands:
    
	ping <ipv6 address, hostname, or routing path>       sends a cjdns ping to the specified node
	route <ipv6 address, hostname, or routing path>      prints out all routes to an IP or the IP to a route
	traceroute <ipv6 address, hostname, or routing path> [-t timeout] performs a traceroute by pinging each known hop to the target on all known paths
	ip <cjdns public key>                                converts a cjdns public key to the corresponding IPv6 address
	host <ipv6 address or hostname>                      returns a list of all know IP address for the specified hostname or the hostname for an address
	cjdnsadmin <-file>                                   creates a .cjdnsadmin file in your home directory using the specified cjdroute.conf as input
	addpeer [-file] [-outfile] '<json peer details>'     adds the peer details to your config file
	addpass [-file] [-outfile] [password]                adds the password to the config if one was supplied, or generates one and then adds
	cleanconfig [-file] [-outfile]                       strips all comments from the config file and then saves it nicely formatted
	log [-l level] [-logfile file] [-line line]          prints cjdns log to stdout
	passgen                                              generates a random alphanumeric password between 15 and 50 characters in length
	peers                                                displays a list of currently connected peers
	dump                                                 dumps the routing table to stdout
	kill                                                 tells cjdns to gracefully exit
	memory                                               returns the number of bytes of memory the router has allocated

**NOTE:** if you don't specify the admin password in the flags then cjdcmd uses the cjdns configuration file to load the details needed to connect. It expects the file to be at `/etc/cjdroute.conf` however you can specify an alternate location with the `-f` or `--file` flags.

### Flags
	
	-c=0: [ping][traceroute] specify the number of packets to send (shorthand)
	-count=0: [ping][traceroute] specify the number of packets to send
	-f="/etc/cjdroute.conf": [all] the cjdroute.conf configuration file to use, edit, or view (shorthand)
	-file="/etc/cjdroute.conf": [all] the cjdroute.conf configuration file to use, edit, or view
	-i=1: [ping] specify the delay between successive pings (shorthand)
	-interval=1: [ping] specify the delay between successive pings
	-l="DEBUG": [log] specify the logging level to use (shorthand)
	-level="DEBUG": [log] specify the logging level to use
	-line=0: [log] specify the cjdns source file line to log
	-logfile="": [log] specify the cjdns source file you wish to see log output from
	-o="/etc/cjdroute.conf": [all] the cjdroute.conf configuration file to save to (shorthand)
	-outfile="/etc/cjdroute.conf": [all] the cjdroute.conf configuration file to save to
	-p="": [all] specify the admin password (shorthand)
	-pass="": [all] specify the admin password
	-t=5000: [ping][traceroute] specify the time in milliseconds cjdns should wait for a response (shorthand)
	-timeout=5000: [ping][traceroute] specify the time in milliseconds cjdns should wait for a response

### Ping

Ping will send a cjdns ping packet to the specified node. Note that this is not the same as an ICMP ping packet like that which is sent with the `ping` and `ping6` utilities; it is a special cjdns switch-level packet. The node will reply with it's version which is the SHA1 hash of the git commit it was built on. 

#### Sample Output

	$ cjdcmd ping -c 4 -t 800 fcf9:11b1:c252:6176:0550:0c59:2bb5:229a
	PING fcf9:11b1:c252:6176:0550:0c59:2bb5:229a (fcf9:11b1:c252:6176:0550:0c59:2bb5:229a)
	Reply from fcf9:11b1:c252:6176:0550:0c59:2bb5:229a 721ms
	Timeout from fcf9:11b1:c252:6176:0550:0c59:2bb5:229a after 830ms
	Reply from fcf9:11b1:c252:6176:0550:0c59:2bb5:229a 771ms
	Reply from fcf9:11b1:c252:6176:0550:0c59:2bb5:229a 778ms
	
	--- fcf9:11b1:c252:6176:0550:0c59:2bb5:229a ping statistics ---
	4 packets transmitted, 3 received, 25.00% packet loss, time 2270ms
	rtt min/avg/max/mdev = 721.000/756.667/778.000/25.382 ms
	Target is using cjdns version d2b1f95ebf39411c37430b7adbd2e76bcc3ad6b6

### Route

Route will either print out all known routes to a specified IPv6 address or the IPv6 address of the node at the end of a specified path. It also displays a human-readable representation of cjdns link quality, which is what the router uses to determine which specific path to take.

Route is useful because it shows you the many different paths that are available to get to one end node. You can then take each of those paths and use the ping command to check for connectivity. This will be further exapnded upon in the future with a `traceroute` command.

#### Sample Output

With an IPv6 address:

	$ cjdcmd route fcf9:11b1:c252:6176:0550:0c59:2bb5:229a
	IP: fcf9:11b1:c252:6176:0550:0c59:2bb5:229a -- Version: 0 -- Path: 0000.0000.0000.f969 -- Link: 0
	IP: fcf9:11b1:c252:6176:0550:0c59:2bb5:229a -- Version: 0 -- Path: 0000.0000.0000.46cd -- Link: 0
	IP: fcf9:11b1:c252:6176:0550:0c59:2bb5:229a -- Version: 1 -- Path: 0000.0000.0000.001f -- Link: 458
	
Or with a path:	

	$ cjdcmd route 0000.0000.0000.001f
	IP: fcf9:11b1:c252:6176:0550:0c59:2bb5:229a -- Version: 1 -- Path: 0000.0000.0000.001f -- Link: 400

### Traceroute

Traceroute will take all the possible routes to a specific target and then ping each known hop along the way. This will display exactly the path your packets take through the network along any given path.

#### Sample Output

	$ cjdcmd traceroute fcf9:11b1:c252:6176:0550:0c59:2bb5:229a
	Finding all routes to fcf9:11b1:c252:6176:0550:0c59:2bb5:229a
	
	Route #1 to target
	IP: fc72:7d84:bac7:3ac2:60cb:e1b3:9025:7266 -- Version: 1 -- Path: 0000.0000.0000.0001 -- Link: 800 -- Time: Skipping ourself
	IP: fc2e:c969:bc94:e8e1:bcef:d155:c13b:ff9f -- Version: 1 -- Path: 0000.0000.0000.00a2 -- Link: 400 -- Time: 1ms 1ms 3ms
	IP: fcf9:11b1:c252:6176:0550:0c59:2bb5:229a -- Version: 0 -- Path: 0000.0000.0000.4d22 -- Link: 303 -- Time: 757ms 815ms 797ms
	
	Route #2 to target
	IP: fc72:7d84:bac7:3ac2:60cb:e1b3:9025:7266 -- Version: 1 -- Path: 0000.0000.0000.0001 -- Link: 800 -- Time: Skipping ourself
	IP: fcef:c7a9:792a:45b3:741f:59aa:9adf:4081 -- Version: 1 -- Path: 0000.0000.0000.0019 -- Link: 402 -- Time: 771ms 781ms 794ms
	IP: fc2e:c969:bc94:e8e1:bcef:d155:c13b:ff9f -- Version: 0 -- Path: 0000.0000.0000.0199 -- Link: 174 -- Time: 1600ms 1654ms 1504ms
	IP: fcf9:11b1:c252:6176:0550:0c59:2bb5:229a -- Version: 1 -- Path: 0000.0000.0000.1f99 -- Link: 136 -- Time: 2367ms 2284ms 2316ms
	
	Route #3 to target
	IP: fc72:7d84:bac7:3ac2:60cb:e1b3:9025:7266 -- Version: 1 -- Path: 0000.0000.0000.0001 -- Link: 800 -- Time: Skipping ourself
	IP: fcf9:11b1:c252:6176:0550:0c59:2bb5:229a -- Version: 1 -- Path: 0000.0000.0000.001f -- Link: 366 -- Time: 796ms 844ms 803ms
	Found 3 routes

### Ip

IP converts a cjdns public key to the matching cjdns IPv6 address. This is useful for editing your peer details since oftentimes all that you will have is the public key.

#### Sample Output:

	$ cjdcmd ip r6jzx210usqbgnm3pdtm1z6btd14pvdtkn5j8qnpgqzknpggkuw0.k
	fc68:cb2c:60db:cb96:19ac:34a8:fd34:03fc


### Host

Host will lookup the cjdns IPv6 address for the given hostname, or will return the hostname for a given IPv6 address. It first tries using your default DNS settings and if no results are found will attempt to use HypeDNS.

#### Sample Output:

	$ cjdcmd host nodeinfo.hype
	nodeinfo.hype has IPv6 address fc5d:baa5:61fc:6ffd:9554:67f0:e290:7535

	$ cjdcmd host fc5d:baa5:61fc:6ffd:9554:67f0:e290:7535
	nodeinfo.hype
	
### Cjdnsadmin

This command will generate a .cjdnsadmin file based on the cjdroute.conf file given to it in the --file flag. If no file is given, it will try using the one specified in ~/.cjdnsadmin. The file contains details on how to connect to a running cjdns instance, as well as your preferred default configuration file. It will be saved as ".cjdnsadmin" in your home directory.

### Addpeer

Addpeer accepts a set of JSON peering details surrounded by ' ' and will walk you through adding them to your config, along with any additional information you would like to save with it. You can specify which file to read and which file to save to using the -file and -outfile flags, both of which are optional. 
 

### Addpass

Addpass optionally accepts a password, or will generate one if none was supplied, and saves it to your config along with any additional information you may wish to add. You can specify which file to read and which file to save to using the -file and -outfile flags, both of which are optional. 


### Cleanconfig

Cleanconfig will read your configuration file, strip the comments, and save it back nicely formatted. You can specify which file to read and which file to save to using the -file and -outfile flags, both of which are optional. 


### Log

Log will begin outputting log information from cjdns. You can optionally specify which level of information to receive which is either Debug, Info, Warn, Error,  or Critical. It also allows you to filter by a specific source code file or a specific line number from the source code. You can use any combination of these options to get the output that you desire. 

#### Sample Output:

	$ cjdcmd log
	1 1357729794 DEBUG Ducttape.c:347 Got running session ver[1] send[12] recv[7] ip[fcd6:b2a5:e3cc:d78d:fc69:a90f:4bf7:4a02]
	2 1357729794 DEBUG SearchStore.c:351 Received response in 781 milliseconds, gmrt now 1035
	3 1357729794 DEBUG Ducttape.c:347 Sending protocol 0 message ver[0] send[2] recv[25] ip[fc6a:d815:ee3b:9bf8:f380:3e58:bc44:2a77]
	4 1357729795 DEBUG RouterModule.c:1137 Ping fc5d:baa5:61fc:6ffd:9554:67f0:e290:7535@0000.0004.fccf.025f
	5 1357729795 DEBUG CryptoAuth.c:568 No traffic in [76] seconds, resetting connection.

### Passgen

Passgen will generate a random alphanumeric password between 15 and 50 characters long.

#### Sample Output:

	$ cjdcmd passgen
	4hVIvpsqkQOTmwY7BdzwQXJe7RfDa3m2tNwulhoTF3K5


### Peers

Peers will list the peers you are directly connected to. This will show both other nodes you connect with and nodes that connect to you. 

#### Sample Output

	$ cjdcmd peers
	Finding all connected peers
	IP: fc8e:753b:2e7f:2575:c895:80d1:d67d:0000 -- Path: 0000.0000.0000.0017 -- Link: 764
	IP: fcd6:b2a5:e3cc:d78d:fc69:a90f:4bf7:4a02 -- Path: 0000.0000.0000.001b -- Link: 764
	IP: fcef:c7a9:792a:45b3:741f:59aa:9adf:4081 -- Path: 0000.0000.0000.0019 -- Link: 484
	IP: fc99:02f4:7795:c86c:36bd:63ae:cf49:d459 -- Path: 0000.0000.0000.001d -- Link: 454
	IP: fc2e:c969:bc94:e8e1:bcef:d155:c13b:ff9f -- Path: 0000.0000.0000.00a2 -- Link: 400
	IP: fcf9:11b1:c252:6176:0550:0c59:2bb5:229a -- Path: 0000.0000.0000.001f -- Link: 400

### Dump

Dump will print the routing table to stdout, complete with IPv6 of the target node, the path to that node (which can be used with the ping command), and human-readable cjdns link quality. Only working paths with a link quality greater than 0 are shown.

#### Sample Output:

	$ cjdcmd dump
	1 IP: fc72:7d84:bac7:3ac2:60cb:e1b3:9025:7266 -- Version: 1 -- Path: 0000.0000.0000.0001 -- Link: 800
	2 IP: fcd8:b768:9762:9808:3d3c:5cac:344c:5261 -- Version: 1 -- Path: 0000.0000.0053.4aad -- Link: 434
	3 IP: fcb1:4025:8840:cf76:c4b1:3202:4d96:c100 -- Version: 0 -- Path: 0000.0000.0000.0a67 -- Link: 431
	...
	142 IP: fcc8:8bbc:51ae:3dbb:bce9:12a1:e563:3b8a -- Version: 1 -- Path: 0000.0000.0000.e67f -- Link: 10
	143 IP: fc66:dfa4:30e8:1844:b0b8:e26e:f120:8fc8 -- Version: 1 -- Path: 0000.aa64.f94b.6eab -- Link: 9
	144 IP: fc1e:af9f:b436:7aa0:5bce:0dfc:0cba:c713 -- Version: 1 -- Path: 0000.8a0a.f94b.6eab -- Link: 9

### Kill

Kill will tell cjdns to shutdown and exit. 

#### Sample Output:

	$ cjdcmd kill
	cjdns is shutting down...


### Memory

Memory returns the number of bytes of memory the router has allocated. 

#### Sample Output:

	$ cjdcmd memory
	90634 bytes
		

Troubleshooting
---------------

### Connection Problems

If you are usign an older version of cjdns, you may get an error similar to this:

	Unable to connect to cjdns: dial tcp 127.0.0.1:11234: connection refused
	dial tcp 127.0.0.1:11234: connection refused
	
Then you either specified the wrong location to connect to or your cjdns version is out of date. Check your configuration file and make sure you are using the latest version of cjdns. 


### Config File

The default configuration file format has changed numerous times, and as such cjdcmd will probably throw an error if you try to use it with an older version of cjdns. Cjdns also allows the use of grossly out of spec JSON without throwing an error which is fine for cjdns but when other programs try to parse the file it causes issues. 

The current format that cjdcmd supports can be found the [configuration file guide](https://github.com/cjdelisle/cjdns/blob/master/rfcs/configure.md) at the cjdns github. The notable changes between what you might have and what is expected are:

* Ensure there are `[` and `]` in the `"UDPInterface"` and `"ETHInterface"` sections, which you can find an example of [here](https://github.com/cjdelisle/cjdns/blob/master/rfcs/configure.md#connection-interfaces.)
* Ensure there are commas after each `{"password":"abcdefghijklmnopqrstuvwxyz"}` section, except the last, like [here](https://github.com/cjdelisle/cjdns/blob/master/rfcs/configure.md#incoming-connections) where they are commented out. For example:


		{"password":"abcdefghijklmnopqrstuvwxyz"},
		{"password":"ABCDEFGHIJKLMNOPQRSTUVWXYZ"},
		{"password":"012345678901234567890123456789"}

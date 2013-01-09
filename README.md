cjdcmd
======

cjdcmd is a command line tool for interfacing with cjdns. It currently supports only very basic functions however more features are currently in development.

Installation
------------

cjdcmd is written in Go. If you already have Go installed, skip the next section.

### Install Go

To install go, check out the [install instructions](http://golang.org/doc/install) or, if you're lucky, you can do a quick install with the information below:

#### Ubuntu (and other Debian variants)

Ubuntu and other Debian variants can use apt to install:

	apt-get install golang

#### Mac OSX

If you have [Homebrew](http://mxcl.github.com/homebrew/) installed:

    brew install go

If your system isn't listed above, perhaps you can [download pre-compiled binaries](http://code.google.com/p/go/downloads), otherwise you will have to [install from source](http://golang.org/doc/install/source).

#### Configure your folders (optional)

Next you should set up a special directory for all your go sourcecode and compiled programs. This prevents you from having to use `sudo` every time you need to install an updated package or program. This is completely optional, however. I will be giving a shortened version of the information found on [the official site](http://golang.org/doc/code.html#tmp_2)

First, make the folder where you want everything to be stored. I use the /home/inhies/projects/go but you may use whatever you like:

    $ mkdir -p $HOME/projects/go 

Next we need to tell our system where to look for Go packages and compiled programs. If you changed the folder name in the previous example, then make sure you change them here. You will want to add this to your `~/.profile` file or similar. On Ubuntu 12.10, I had to add it to my `~/.bashrc`:

	export GOPATH=$HOME/projects/go
	export PATH=$PATH:$HOME/projects/go/bin

Now, to make the changes take effect immediately, you can either run `$ source ~/.bashrc` or just paste those two lines you added to the file on your command line. You should now be setup! Try typing `$ echo $GOPATH` and make sure you see the folder you specified. If you have any problems, try re-reading the [official documentation](http://golang.org/doc/code.html#tmp_2).

### Install cjdcmd

Once you have Go installed, installing new programs and packages couldn't be easier. Simply run the following command to have cjdcmd download, build, and install:

    go get github.com/inhies/cjdcmd
	
**NOTE:** You may have to be root (use `sudo`) to install Go and cjdcmd.
	
Using cjdcmd
------------

Once you have cjdcmd installed you can run it without any arguments to get a list of commands or run it with the flag `--help` to get a list of all support options. Currently cjdcmd offers the following commands:
    
	ping  <cjdns IPv6 address or cjdns path (obtained with route command)>
	route <cjdns IPv6 address or cjdns path (obtained with route command)>
	log
	dump
	kill

**NOTE:** cjdcmd uses the cjdns configuration file to load the details needed to connect. It expects the file to be at `/etc/cjdroute.conf` however you can specify an alternate location with the `-f` or `--file` flags.

### Ping

Ping will send a cjdns ping packet to the specified node. Note that this is not the same as an ICMP ping packet like that which is sent with the `ping` and `ping6` utilities; it is a special cjdns switch-level packet. The node will reply with it's version which is the SHA1 hash of the git commit it was built on. 

#### Sample Output:

	$cjdcmd ping -c 4 -t 800 fcf9:11b1:c252:6176:0550:0c59:2bb5:229a
	Attempting to connect to cjdns...Connected
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

#### Sample Output:

With an IPv6 address:

	$cjdcmd route fcf9:11b1:c252:6176:0550:0c59:2bb5:229a
	Attempting to connect to cjdns...Connected
	IP: fcf9:11b1:c252:6176:0550:0c59:2bb5:229a -- Version: 0 -- Path: 0000.0000.0000.f969 -- Link: 0
	IP: fcf9:11b1:c252:6176:0550:0c59:2bb5:229a -- Version: 0 -- Path: 0000.0000.0000.46cd -- Link: 0
	IP: fcf9:11b1:c252:6176:0550:0c59:2bb5:229a -- Version: 1 -- Path: 0000.0000.0000.001f -- Link: 458
	
Or with a path:	

	$cjdcmd route 0000.0000.0000.001f
	Attempting to connect to cjdns...Connected
	IP: fcf9:11b1:c252:6176:0550:0c59:2bb5:229a -- Version: 1 -- Path: 0000.0000.0000.001f -- Link: 400

### Log

Log will begin outputting log information from cjdns. You can optionally specify which level of information to receive which is either Debug, Info, Warn, Error,  or Critical. It also allows you to filter by a specific source code file or a specific line number from the source code. You can use any combination of these options to get the output that you desire. 

#### Sample Output:

	$cjdcmd log
	Attempting to connect to cjdns...Connected
	1 1357729794 DEBUG Ducttape.c:347 Got running session ver[1] send[12] recv[7] ip[fcd6:b2a5:e3cc:d78d:fc69:a90f:4bf7:4a02]
	2 1357729794 DEBUG SearchStore.c:351 Received response in 781 milliseconds, gmrt now 1035
	3 1357729794 DEBUG Ducttape.c:347 Sending protocol 0 message ver[0] send[2] recv[25] ip[fc6a:d815:ee3b:9bf8:f380:3e58:bc44:2a77]
	4 1357729795 DEBUG RouterModule.c:1137 Ping fc5d:baa5:61fc:6ffd:9554:67f0:e290:7535@0000.0004.fccf.025f
	5 1357729795 DEBUG CryptoAuth.c:568 No traffic in [76] seconds, resetting connection.


### Dump

Dump will print the routing table to stdout, complete with IPv6 of the target node, the path to that node (which can be used with the ping command), and human-readable cjdns link quality. Note that a quality of 0 means the path is dead and should not be used.

#### Sample Output:

	$cjdcmd dump
	Attempting to connect to cjdns...Connected
	0 IP: fc72:7d84:bac7:3ac2:60cb:e1b3:9025:7266 -- Version: 1 -- Path: 0000.0000.0000.0001 -- Link: 800
	1 IP: fcd9:6a75:6c9c:65dd:318f:26f0:1319:d0d3 -- Version: 1 -- Path: 0000.0000.0000.0b6b -- Link: 367
	2 IP: fcda:947a:ee48:802f:65ed:7e8a:2fb6:bf6b -- Version: 1 -- Path: 0000.0000.2894.c80b -- Link: 0
	3 IP: fcf2:4de8:ffa1:9679:821a:b0c9:961f:5b98 -- Version: 1 -- Path: 0000.0000.0059.186b -- Link: 363
	...
	487 IP: fcac:541e:9c5c:9ddc:f648:962a:2892:e33e -- Version: 0 -- Path: 0000.011d.1f72.b25f -- Link: 0
	488 IP: fc21:103b:fbc8:828c:810c:37b6:3b1e:9615 -- Version: 0 -- Path: 0000.0004.cf72.b25f -- Link: 0
	489 IP: fc74:b146:a580:2be9:6285:7af3:6a56:2b7b -- Version: 1 -- Path: 0000.0000.0757.1eab -- Link: 0
	490 IP: fca5:9fe0:3fa2:d576:71e6:8373:7aeb:ea11 -- Version: 1 -- Path: 0000.0000.0000.484b -- Link: 0
	
### Kill

Kill will tell cjdns to shutdown and exit. 

#### Sample Output:

	$cjdcmd kill
	Attempting to connect to cjdns...Connected
	cjdns is shutting down...
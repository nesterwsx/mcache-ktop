

# mcache-ktop

mcache-ktop is a libpcap based tool to interactively show the top of memcached keys used.

 - supports keys collapsing by regexp
 - sorting by most readable/writable keys
 
Inspired by https://github.com/reddit/mcsauna and https://github.com/etsy/mctop

## Installation

CentOS: 
```
yum -y install libpcap-devel ncurses-devel
git clone https://github.com/inperfo/mcache-ktop && cd mcache-ktop
grep github main.go | xargs go get
go build -o mcache-ktop *.go && ./mcache-ktop -i eth0 -d 1 
```

Tested on CentOS x64 6/7 and MacOS 

## Example
```
       Reads	   Writes        Read bytes	      Written bytes   Key
---------------------------------------------------------------------------------------------------------
   3,168,060       16,774            11 GB              140 MB   user_*
   1,718,936           75           1.4 GB               91 MB   yi:SHOW_CREATE_TABLE_*
   1,576,521           58           2.3 GB               89 MB   yi:SHOW_FULL_COLUMNS_FROM*
   1,185,086        7,755           2.9 GB               41 MB   hrc*
     631,177            0           8.9 GB               11 MB   m_rus:*
     454,220       31,501           180 MB               42 MB   photo_id_*
     444,676            1            27 MB               14 MB   settings:cache_version
     414,389            0              0 B              8.7 MB   banlist:
     279,263       11,324            21 MB               17 MB   Vcl_*
     217,824            0           1.6 GB              6.4 MB   map_coords:*
     195,357        2,275            18 MB               13 MB   history_*
     150,032            0           778 MB              3.6 MB   ads_limit_*
      97,768            0           144 MB              4.9 MB   yi:default:`region`.`smart_blocks`
      56,492            0            39 MB              2.5 MB   yi:default:`log`.`announcement`
      53,023            0            75 MB              1.4 MB   sbdm_by_name:.ru
      52,019            0            76 MB              2.3 MB   meta_tags_rules:site_page_meta_tags
      47,084            0            67 MB              1.2 MB   sbdm_by_name:spb:ru
      40,150            4           7.4 MB              2.2 MB   meta_tags:8d3e8bc45df1f292a4a701e50921eb
      37,795            1            50 MB              1.4 MB   metro_groups:station_to_district
      36,376            0            12 MB              946 kB   sbdm_by_name:api:.ru
      26,163            0            37 MB              1.3 MB   words_replacements:words_replacements_ar
      25,141            0           9.4 MB              830 kB   exp:2073_step1
      24,167            0            30 MB              1.1 MB   meta_tags_rules:site_page_meta_tags_15
      22,694            1           183 MB              1.8 MB   LinksMEMCacheEngine:734950d42d8cd98f00b2
      21,599            0            27 MB              1.1 MB   yi:default:`log`.`payment_informer`
      18,406            0            19 MB              626 kB   search_region_cache:e1cbda03
      18,002           24            11 MB              1.0 MB   apps:mobile_vrs_970482efa9827e28dce2589c
```      

## Command line options

```
Usage mcache-ktop [options]:

  -c string
        config file
  -d int
        update interval (seconds, default 3)
  -h string
        capture ip address (i.e. for bond with multiple IPs)
  -i string
        capture interface (default "any")
  -p int
        capture port (default 11211)
  -w string
        file to write output to
```

### Example of config file
```
{
        "regexps": [
                {"re": "^user_:.*",     "name": "user_*"},
                {"re": "^hrc:.*",       "name": "hrc*"},
		        {"re": "^Vlc_.*",       "name": "Vlc_*"},
		        {"re": "^history_.*",   "name": "history_*"},
		        {"re": "^map_coords.*", "name": "map_coords*"}
        ],
        "interval": 3,
        "interface": "",
        "ip": "",
        "port": 11211,
        "quiet": false,
	"output_file": ""
}
```
## Known issues

### gopacket/pcap sometimes crash app
.

### No binary protocol support
There is currently no support for the binary protocol. However, if someone is using it and would like to submit a patch, it would be welcome.


# flustro
Flustro (whisper in esperanto) is a toolkit for dealing with whisper files.
---
[![Build Status](https://travis-ci.org/fuzzy/flustro.svg)](https://travis-ci.org/fuzzy/flustro) [![Codebeat](https://codebeat.co/badges/06369d13-abb0-4e16-946b-f13c242bd88c)](https://codebeat.co/a/fuzzy-wombat-iii/projects/flustro-master)

## Installation
```
$ go get -u github.com/fuzzy/flustro
```

## Usage Examples

#### flustro dump
```
$ flustro dump whisper/collectd/hostname/cpu-0/user.wsp
>> File           | AggMethod      | MaxRetention   | NumArchives                                                  
>> system_pct.wsp | average        | 94608000       | 5                                                            

>> Archive        | Offset         | NumPoints      | Interval       | Retention      | Size                       
>> 0              | 76             | 60480          | 10             | 604800         | 725760                     
>> 1              | 725836         | 20160          | 60             | 1209600        | 241920                     
>> 2              | 967756         | 8064           | 300            | 2419200        | 96768                      
>> 3              | 1064524        | 2016           | 3600           | 7257600        | 24192                      
>> 4              | 1088716        | 1095           | 86400          | 94608000       | 13140

$ flustro dump -P whisper/collectd/hostname/cpu-0/user.wsp
>> File           | AggMethod      | MaxRetention   | NumArchives                                                  
>> system_pct.wsp | average        | 94608000       | 5                                                            

>> Archive        | Offset         | NumPoints      | Interval       | Retention      | Size                       
>> 0              | 76             | 60480          | 10             | 604800         | 725760                     
>> Timestamp       | Value                                                                                         
>> 1471018640      | 10.199789695057834                                                                            
>> 1471018650      | 10.031023784901757
...

$ flustro dump -P -A 2 whisper/collectd/hostname/cpu-0/user.wsp
>> File           | AggMethod      | MaxRetention   | NumArchives                                                  
>> system_pct.wsp | average        | 94608000       | 5                                                            

>> Archive        | Offset         | NumPoints      | Interval       | Retention      | Size                       
>> 2              | 76             | 60480          | 10             | 604800         | 725760                     
>> Timestamp       | Value                                                                                         
>> 1471018620      | 10.298661174047373                                                                            
....
```

#### flustro fill / flustro merge

The fill and merge commands use the same backend functions, and work in the same way, with the only difference
being that merge will overwrite existing datapoints, where fill will only write values if the destination is nan.

```
$ flustro help fill
NAME:
   flustro fill - Backfill datapoints in the dst from the src

USAGE:
   flustro fill [command options] <src(File|Dir)> <dst(File|Dir)>

DESCRIPTION:
   Backfill datapoints in the dst from the src

OPTIONS:
   -J value  Number of workers (for directory recursion) 
```

You can use this to backfill a single whisper file

```
$ flustro fill whisper/collectd/hostname/cpu-0/user.wsp whisper2/collectd/hostname/cpu-0/user.wsp
```

Or an entire directory of whisper files
```
$ flustro fill whisper/collectd/hostname whisper2/collectd/hostname
```

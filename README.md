#### synctool

This tool is being written to help me with some arbitrary URL based file downloads, as well as a good excuse to experement with some different features in golang.


What it does:

* Read in a `--file` `-f` `FILE` that should be newline separated URL's of files to download
* Iterate over those url's line by line and `--output` `-o` to an `./output` directory

There is an example filelist.txt with newline separated URL's showing the example of what could feasably be downloaded with this utility.

# TODO: 

* Handle failure
  * Retry with exponential backoff
* Maybe replace https://github.com/BadgerOps/epel-offline-sync ?

# Notes:

* `--loglevel value` is added to support logrus based log levels.
* I am using [urfave/cli v2](https://cli.urfave.org/v2/) because I wanted to experiment with a cli tool

# Example output:

```bash
❯ go run main.go -f filelist.txt
INFO[2024-04-03 16:48:36] Downloading file from URL: https://blog.badgerops.net/content/images/2020/03/badger.png 
INFO[2024-04-03 16:48:36] Downloading file from URL: https://ash-speed.hetzner.com/1GB.bin 
INFO[2024-04-03 16:48:36] Downloading file from URL: https://www.bobrossquotes.com/bobs/bob.png 
INFO[2024-04-03 16:48:36] Downloading file from URL: https://ash-speed.hetzner.com/100MB.bin 
INFO[2024-04-03 16:48:36] Total download time for url https://blog.badgerops.net/content/images/2020/03/badger.png: 0s 
INFO[2024-04-03 16:48:36] Total file size for url https://blog.badgerops.net/content/images/2020/03/badger.png: 0 B 
INFO[2024-04-03 16:48:36] Downloading file from URL: https://download.opensuse.org/distribution/leap/15.0/iso/openSUSE-Leap-15.0-NET-x86_64.iso 
INFO[2024-04-03 16:48:37] Total download time for url https://www.bobrossquotes.com/bobs/bob.png: 0s 
INFO[2024-04-03 16:48:37] Total file size for url https://www.bobrossquotes.com/bobs/bob.png: 0 B 
INFO[2024-04-03 16:48:41] Current download rate for: https://ash-speed.hetzner.com/1GB.bin is 3.8 MiB/s  threadID=2
INFO[2024-04-03 16:48:41] Current download rate for: https://ash-speed.hetzner.com/100MB.bin is 3.3 MiB/s  threadID=0
INFO[2024-04-03 16:48:41] Current download rate for: https://download.opensuse.org/distribution/leap/15.0/iso/openSUSE-Leap-15.0-NET-x86_64.iso is 23.2 MiB/s  threadID=3
INFO[2024-04-03 16:48:42] Total download time for url https://download.opensuse.org/distribution/leap/15.0/iso/openSUSE-Leap-15.0-NET-x86_64.iso: 5s 
INFO[2024-04-03 16:48:42] Total file size for url https://download.opensuse.org/distribution/leap/15.0/iso/openSUSE-Leap-15.0-NET-x86_64.iso: 0 B 
INFO[2024-04-03 16:48:46] Current download rate for: https://ash-speed.hetzner.com/1GB.bin is 4.2 MiB/s  threadID=2
INFO[2024-04-03 16:48:46] Current download rate for: https://ash-speed.hetzner.com/100MB.bin is 3.6 MiB/s  threadID=0
INFO[2024-04-03 16:48:51] Current download rate for: https://ash-speed.hetzner.com/1GB.bin is 4.4 MiB/s  threadID=2
INFO[2024-04-03 16:48:51] Current download rate for: https://ash-speed.hetzner.com/100MB.bin is 3.8 MiB/s  threadID=0
INFO[2024-04-03 16:48:56] Current download rate for: https://ash-speed.hetzner.com/1GB.bin is 5.3 MiB/s  threadID=2
INFO[2024-04-03 16:48:56] Current download rate for: https://ash-speed.hetzner.com/100MB.bin is 4.6 MiB/s  threadID=0
INFO[2024-04-03 16:48:57] Total download time for url https://ash-speed.hetzner.com/100MB.bin: 20s 
```

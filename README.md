#### synctool

This tool is being written to help me with some arbitrary URL based file downloads, as well as a good excuse to experement with some different features in golang.


What it does:

* Read in a `--file` `-f` `FILE` that should be newline separated URL's of files to download
* Iterate over those url's line by line and `--output` `-o` to an `./output` directory

There is an example filelist.txt with newline separated URL's showing the example of what could feasably be downloaded with this utility.

#TODO: 

* Handle failure
  * Retry with exponential backoff
* Maybe replace https://github.com/BadgerOps/epel-offline-sync ?

# Notes:

* `--loglevel value` is added to support logrus based log levels.
* I am using [urfave/cli v2](https://cli.urfave.org/v2/) because I wanted to experiment with a cli tool
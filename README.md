#### synctool

I'm writing this tool mostly as a way to explore golang. Don't use it!

What it does:

* Read in a `--file` `-f` `FILE` that should be newline separated URL's of files to download
* Iterate over those url's line by line and `--output` `-o` to an `./output` directory

#TODO: 

* Handle failure
* Retry with exponential backoff
* Maybe replace https://github.com/BadgerOps/epel-offline-sync ?

# Notes:

* I used `ollama` and `mistral:7b` inside vscode with [continue.dev](https://continue.dev) while writing this
  * It was often wrong, so I also ended up just _reading the docs_
* I used [urfave/cli v2](https://cli.urfave.org/v2/) because I wanted to experiment with a cli tool

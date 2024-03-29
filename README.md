# GitHunter
Every now and then, on an internal test, you'll land on a box with a Git repository already checked out into directory. GitHunter will help you search that repository for anything juicy which could help you with other areas of your test.

GitHunter looks for:
* Keywords in commit messages
* Keywords in files
* Interesting file names

Thanks to [@michenriksen](https://github.com/michenriksen/) for allowing me to include his existing [Gitrob](https://github.com/michenriksen/gitrob) signatures for doing the file name checks.

For the two keyword checks, the script uses a customisable JSON file to allow you to do either simple or regular expressesion searches, meaning you can target the discovery to your client's environment.

## Installation
These instructions aren't best practice, ideally you would install this with `go install...` but as far as I can tell this will only install the binary. I also need to install the `patterns.json` file somewhere so it can be used by the app and edited by users, so this is the best I can come up with for now. Any suggestions, let me know.

1. [Set up your Go environment.](https://golang.org/doc/install)
2. Change into the Go source directory:
   ```
   cd ~go/src/github.com
   ```
1. Make a directory for me and go into it:
   ```
   mkdir digininja ; cd digininja
   ```
1. Clone the project:
   ```
   git clone https://github.com/digininja/GitHunter.git
   ```
1. Change into the project directory:
   ```
   cd GitHunter/
   ```
1. Build it:
   ```
   go build
   ```

You should now have a binary in the current directory.

## Usage
Usage is fairly simple, by default, GitHunter will look in the current directory for a `.git` directory and, if it finds one, will parse through it and show anything interesting it finds in either filenames or in commit comments. You can specify a different directory for the repository with the `-gitdir` parameter.

If you want to expand what is searched to include file contents at each commit, you can add the `-grep` parameter, but be warned, git, on my box at least, runs single threaded, and can take a long time to do the grepping on a large repository. It actually failed trying to grep through Metasploit, due to the sheer number of commits and content. Still worth trying it though, especially on smaller repos, as you may find something.

If you want a dump of the commit logs, without any commentary, then you can use the `-dump` parameter.

To specify a custom patterns file, use `-patterns` and to have the output without any fancy colours (easier for parsing) use `-nocolours`.

## Testing things out
If you want a repository to test things on, have a look at my [Leaky Repo](https://github.com/digininja/leakyrepo) which contains quite a few interesting things to find.


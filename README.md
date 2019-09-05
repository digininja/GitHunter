# GitHunter
Every now and then, on an internal test, you'll land on a box with a Git repository already checked out into directory. GitHunter will help you search that repository for anything juicy which could help you with other areas of your test.

GitHunter looks for:
* Keywords in commit messages
* Keywords in files
* Interesting file names

Thanks to [@michenriksen](https://github.com/michenriksen/) for allowing me to include his existing [Gitrob](https://github.com/michenriksen/gitrob) signatures for doing the file name checks.

For the two keyword checks, the script uses a customisable JSON file to allow you to do either simple or regular expressesion searches, meaning you can target the discovery to your client's environment.

## Installation
1. [Set up your Go environment.](https://golang.org/doc/install)
1. Checkout the project:
    ```
    go get https://github.com/digininja/GitHunter/
    ```
1. Get any dependencies:
    ```
    go get ./...
    ```
    
## Usage
Usage is fairly simple, by default, GitHunter will look in the current directory for a `.git` directory and, if it finds one, will parse through it and show anything interesting it finds. You can specify a different directory for the repository with the `-gitdir` parameter.

If you want a dump of the commit logs, without any commentary, then you can use the `-dump` parameter.

To specify a custom patterns file, use `-patterns` and to have the output without any fancy colours (easier for parsing) use `-nocolours`.

## Testing things out
If you want a repository to test things on, have a look at my [Leaky Repo](https://github.com/digininja/leakyrepo) which contains quite a few interesting things to find.

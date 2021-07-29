# Contributing to the Flow CLI

The following is a set of guidelines for contributing to the Flow CLI.
These are mostly guidelines, not rules.
Use your best judgment, and feel free to propose changes to this document in a pull request.

## Table Of Contents

[Getting Started](#project-overview)

[How Can I Contribute?](#how-can-i-contribute)

- [Reporting Bugs](#reporting-bugs)
- [Suggesting Enhancements](#suggesting-enhancements)
- [Your First Code Contribution](#your-first-code-contribution)
- [Pull Requests](#pull-requests)

[Styleguides](#styleguides)

- [Git Commit Messages](#git-commit-messages)
- [Go Styleguide](#go-styleguide)

[Additional Notes](#additional-notes)

## How Can I Contribute?

### Reporting Bugs

#### Before Submitting A Bug Report

- **Search existing issues** to see if the problem has already been reported.
  If it has **and the issue is still open**, add a comment to the existing issue instead of opening a new one.

#### How Do I Submit A (Good) Bug Report?

Explain the problem and include additional details to help maintainers reproduce the problem:

- **Use a clear and descriptive title** for the issue to identify the problem.
- **Describe the exact steps which reproduce the problem** in as many details as possible.
  When listing steps, **don't just say what you did, but explain how you did it**.
- **Provide specific examples to demonstrate the steps**.
  Include links to files or GitHub projects, or copy/pasteable snippets, which you use in those examples.
  If you're providing snippets in the issue,
  use [Markdown code blocks](https://help.github.com/articles/markdown-basics/#multiple-lines).
- **Describe the behavior you observed after following the steps** and point out what exactly is the problem with that behavior.
- **Explain which behavior you expected to see instead and why.**
- **Include error messages and stack traces** which show the output / crash and clearly demonstrate the problem.

Provide more context by answering these questions:

- **Can you reliably reproduce the issue?** If not, provide details about how often the problem happens
  and under which conditions it normally happens.

Include details about your configuration and environment:

- **What is the version of the CLI you're using**?
- **What's the name and version of the Operating System you're using**?

### Suggesting Enhancements

#### Before Submitting An Enhancement Suggestion

- **Perform a cursory search** to see if the enhancement has already been suggested.
  If it has, add a comment to the existing issue instead of opening a new one.

#### How Do I Submit A (Good) Enhancement Suggestion?

Enhancement suggestions are tracked as [GitHub issues](https://guides.github.com/features/issues/).
Create an issue and provide the following information:

- **Use a clear and descriptive title** for the issue to identify the suggestion.
- **Provide a step-by-step description of the suggested enhancement** in as many details as possible.
- **Provide specific examples to demonstrate the steps**.
  Include copy/pasteable snippets which you use in those examples,
  as [Markdown code blocks](https://help.github.com/articles/markdown-basics/#multiple-lines).
- **Describe the current behavior** and **explain which behavior you expected to see instead** and why.
- **Explain why this enhancement would be useful** to CLI users.

### Your First Code Contribution

Unsure where to begin contributing to the Flow CLI?
You can start by looking through these `good-first-issue` and `help-wanted` issues:

- [Good first issues](https://github.com/onflow/flow-cli/labels/good%20first%20issue):
  issues which should only require a few lines of code, and a test or two.
- [Help wanted issues](https://github.com/onflow/flow-cli/labels/help%20wanted):
  issues which should be a bit more involved than `good-first-issue` issues.

Both issue lists are sorted by total number of comments.
While not perfect, number of comments is a reasonable proxy for impact a given change will have.

### Pull Requests

The process described here has several goals:

- Maintain code quality
- Fix problems that are important to users
- Engage the community in working toward the best possible UX
- Enable a sustainable system for the CLI's maintainers to review contributions

Please follow the [styleguides](#styleguides) to have your contribution considered by the maintainers.
Reviewer(s) may ask you to complete additional design work, tests,
or other changes before your pull request can be ultimately accepted.

## Styleguides

Before contributing, make sure to examine the project to get familiar with the patterns and style already being used.

### Git Commit Messages

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or less
- Reference issues and pull requests liberally after the first line

### Go Styleguide

The majority of this project is written Go.

We try to follow the coding guidelines from the Go community.

- Code should be formatted using `gofmt`
- Code should pass the linter: `make lint`
- Code should follow the guidelines covered in
  [Effective Go](http://golang.org/doc/effective_go.html)
  and [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Code should be commented
- Code should pass all tests: `make test`

## Releasing

Releasing is automated by Github actions. Release action is triggered by creating a release on Github and publishing it.

You can also release manually although this is not recommended:
- Tag a new release and push it
- Build the binaries: `make versioned-binaries`
- Test built binaries locally
- Upload the binaries: `make publish`
- Update the Homebrew formula: e.g. `brew bump-formula-pr flow-cli --version=0.12.3`

To make the new version the default version that is installed

- Change `version.txt` and commit it
- Upload the version file: `gsutil cp version.txt gs://flow-cli`


# CLI Guidelines
This is a design guideline used for the development of the Flow CLI. The purpose of this guideline is to achieve consistency across new features and allow composability of commands that build the fundamentals of great CLI design.

> Whatever software you’re building, you can be absolutely certain that people will
> use it in ways you didn’t anticipate. Your software will become a part in a larger system
> — your only choice is over whether it will be a well-behaved part. - [Clig](https://clig.dev/)


# Human Interaction

Our interaction with the CLI is done by executing CLI commands and passing arguments
and flags which need to be consistent. Consistency will allow interactions to
be predictable and will with time feel natural even without resorting to reading instructions.

## Commands
Be consistent across commands and subcommands. In case we define a subcommand we
should use `noun verb` pattern. **Don’t have ambiguous or similarly-named commands.**
Naming and language should be the same as used in the rest of Flow's documentation. 
Use only lowercase letters, and dashes if you really need to. Keep it short.
Users will be typing it all the time.

```
flow accounts
```

## Flags
Flags are **named parameters**, denoted with either a hyphen, and a
single-letter name (`-r`) or a double hyphen and a multiple-letter
name (`--recursive`). The longer version is preferred for better readability
and simpler learning experience in the beginning but after a while users will
probably start using the shorter version. 

Every flag should have a shorter version and they should be both presented in the help.
Every flag that can have a default value should have a default value 
so users don’t have to specify all of them when not needed. 

Support flags following a value delimited by a space or
equal (=) sign (ex: `--username test` or `--username=test`)

```
flow accounts get --filter "address"
```

Use common naming for default falgs such as:

`--log`: Logging level.

`--output`: Output format (JSON, inline etc...).

`-h, --help`: Help. This should only mean help. See the help section.

`--version`: Version.


## Arguments
Arguments, or args, are positional parameters to a command. 
There should be as few arguments as possible, because it is hard for users to remember the sequence of arguments.

Because they are relying on position, flags should be used where more than one argument would be required. 

Arguments should be one word verbs following general naming guideline. 
**Prefer flags to arguments**.

Use an argument for the value that command requires in order to run.

```
flow accounts get <address>
```

## Robustness
CLI should feel robust. Long operation should also provide output as they are processing.
**Responsivness is more important than speed.**

Warn before you make a non-additive change. Eventually, you’ll find that you can’t
avoid breaking an interface. Before you do, forewarn your users in the program itself:
when they pass the flag you’re looking to deprecate, tell them it’s going to change soon.
Make sure there’s a way they can modify their usage today to make it
future-proof, and tell them how to do it.

# Computer Interaction

The CLI will respond to execution of commands by providing some output.
Output should with the same reasons as input be consistent and predictable,
even further it should follow some industry standards and thus allow great strength
of CLI use: composability. Output should be served in the right dosage. We shouldn't
output lines and lines of data because we will shadow the important part the same
way we shouldn't let the CLI command run in background not providing any output for
minutes as it is working because the user will start to doubt if the command is broken.
It's important to get this balance right.

## Interaction
Commands must be stateless and idempotent, meaning they can be run without relying on external state.
If we have some external state, like the configuration, each command must include an option to
define it on the fly, but it must also work without it by first relying on the externally
saved state in the configuration, and if not found, relying on a default value.
When relying on a default value results in an error, there should be an explanation for the user that a configuration should be created. 

We try to use default values first to get that “works like magic” feeling.

Never require a prompt for user input. Always offer a flag to provide that input.
However, if a user doesn't provide required input we can offer a prompt as an alternative.


## Output

“Expect the output of every program to become the input to another, as yet unknown, program.” — Doug McIlroy

Output should use **stdout**. Output of commands should be presented in a clear formatted way for
users to easily scan it, but it must also be compatible with `grep` command often used in
command chaining. Output should also be possible in json by using `--output json` flag.

Default command response should be to the stdout and not saved to a file. Anytime we want
the output to be saved to a file we should explicitly specify so by using `--save filename.txt`
flag and providing the path.

Result output should include only information that is commonly used and relevant, 
don't use too much of user screen drowning what’s truly important, 
instead provide a way to include that data when the user requests by having include, exclude flags.

```
Address  179b6b1cb6755e31
Balance  0
Keys     2

Key 0   Public Key               c8a2a318b9099cc6c872a0ec3dcd9f59d17837e4ffd6cd8a1f913ddfa769559605e1ad6ad603ebb511f5a6c8125f863abc2e9c600216edaa07104a0fe320dba7
        Weight                   1000
        Signature Algorithm      ECDSA_P256
        Hash Algorithm           SHA3_256

Code             
         pub contract Foo {
                pub var bar: String
         
                init() {
                        self.bar = "Hello, World!"
                }
         }
```

```
{"Address":"179b6b1cb6755e31","Balance":0,"Code":"CnB1YiBjb250cmFjdCBGb28gewoJcHViIHZhciBiYXI6IFN0cmluZwoKCWluaXQoKSB7CgkJc2VsZi5iYXIgPSAiSGVsbG8sIFdvcmxkISIKCX0KfQo=","Keys":[{"Index":0,"PublicKey":{},"SigAlgo":2,"HashAlgo":3,"Weight":1000,"SequenceNumber":0,"Revoked":false}],"Contracts":null}
```

## Error
Error should be human-readable, avoid printing out stack trace. It should give some
pointers as to what could have gone wrong and how to fix it.
Maybe try involving an example. However allow `--log debug` flag for complete error info
with stack. Error messages should be sent to `stderr`

```
❌ Error while dialing dial tcp 127.0.0.1:3569: connect: connection refused" 
🙏 Make sure your emulator is running or connection address is correct.
```

## Saving
Saving output to files should be optional and never a default action unless it is
clear from the command name the output will be saved and that is
required for later use (example: `flow init`). Output can be piped in files, and
our CLI should resort to that for default saving. If there is a specific format
that needs to be saved or converted to we should be able to display that format as
well hence again allowing us to pipe it to files.

## Help
Main help screen must contain a list of all commands with their own description,
it must also include examples for commands. Help screen should include link to the documentation website.

Commands should have a description that is not too long (less than 100 characters).
Help should be outputed if command is ran without any arguments. Help outline should be:

```Description:
    <description>

Usage:
    <command> <action>

Examples:
    An optional section of example(s) on how to run the command.

Commands:
    <command>
        <commandDescription>
    <command>
        <commandDescription>
    …

Flags:
    --<flag>
        <flagDescription>
    --<flag>
        <flagDescription>
    …
```

## Progress
Show progress for longer running commands visually. Visual feedback on longer
running commands is important so users don’t get confused if command is running
or the CLI is hang up resulting in user quitting.

```
Loading 0x1fd892083b3e2a4c...⠼
```

## Feedback
Commands should provide feedback if there is no result to be presented.
If a command completes an operation that has no result to be shown it
should write out that command was executed. This is meant to assure the user
of the completion of the command.
If user is executing a destructive command they should be asked for approval
but this command should allow passing confirmation with a flag `--yes`.

```
💾 result saved to: account.txt 
```

## Exit
CLI should return zero status unless it is shut down because of an error.
This will allow chaining of commands and interfacing with other cli. If a
command is long running then provide description how to exit it (e.g. "Press Ctrl+C to stop").
```
exit status 1
```

## Colors
Base color should be white with yellow reserved for warnings and red for errors.
Blue color can be used to accent some commands but it shouldn’t be
used too much as it will confuse the user.


## Inspiration and Reference
https://clig.dev/

https://blog.developer.atlassian.com/10-design-principles-for-delightful-clis/

https://devcenter.heroku.com/articles/cli-style-guide

https://eng.localytics.com/exploring-cli-best-practices/










## Additional Notes

Thank you for your interest in contributing to the Flow CLI!

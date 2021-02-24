# CLI Style Guide
This is a design guideline used for development of Flow CLI. Purpose of this guidline is to achieve consistency across new features and allow composability of commands that build the fundaments of great CLI design.

> Whatever software you’re building, you can be absolutely certain that people will use it in ways you didn’t anticipate. Your software will become a part in a larger system—your only choice is over whether it will be a well-behaved part. - [Clig](https://clig.dev/)


# Human Interaction

Our interaction with CLI is done by executing CLI commands and passing arguments and flags which need to be consistent. Consistency will allow interactions to be predictable and will with time feel natural even without resorting to reading instructions.

## Commands
**TBD**
Be consistent across commands and subcomands. In case we define a subcommand we should use `noun verb` pattern. **Don’t have ambiguous or similarly-named commands.**
Naming and language should be same as used in Flow documentation. Use only lowercase letters, and dashes if you really need to. Keep it short. Users will be typing it all the time.

```
flow account
```

## Flags
Flags are **named parameters**, denoted with either a hyphen and a single-letter name (`-r`) or a double hyphen and a multiple-letter name (`--recursive`). Longer version is preferred for better readability and simpler learning experience in the beginning but after users will probably start using the shorter version. Every flag should have a shorter version and they should be both presented in the help. Every flag that can have a default value should have a default value so users don’t have to specify all of them when not needed. Support flags following a value delimited by a space or equal (=) sign (ex: `--username test` or `--username=test`)

```
flow account get --filter "address"
```

Use common naming for default falgs such as:

`-a, --all`: All. For example, ps, fetchmail.

`-d, --debug`: Show debugging output.

`-f, --force`: Force. For example, rm -f will force the removal of files, even if it thinks it does not have permission to do it. This is also useful for commands which are doing something destructive that usually require user confirmation, but you want to force it to do that destructive action in a script.

`--json`: Display JSON output. See the output section.

`-h, --help`: Help. This should only mean help. See the help section.

`--no-input`: See the interactivity section.

`-o, --output`: Output file. For example, sort, gcc.

`-p, --port`: Port. For example, psql, ssh.

`-q, --quiet`: Quiet. Display less output. This is particularly useful when displaying output 
for humans that you might want to hide when running in a script.

`-u, --user`: User. For example, ps, ssh.

`--version`: Version.


## Arguments
Arguments, or args, are positional parameters to a command. There should be as little arguments as possible because they make it hard for users to remember the sequence of arguments. Because they are relaying on position flags should be used where more than one argument would be required. Arguments should be one word verbs following general naming guideline. **Preffer flags to arguments**.

```
flow account get
```

## Robsutness
CLI should feel robust. We can achieve this by thinking of different ways users will malluse the CLI and provide feedback. Long operation should also provide output as they are processing. **Responsivness is more important than speed.**

Warn before you make a non-additive change. Eventually, you’ll find that you can’t avoid breaking an interface. Before you do, forewarn your users in the program itself: when they pass the flag you’re looking to deprecate, tell them it’s going to change soon. Make sure there’s a way they can modify their usage today to make it future-proof, and tell them how to do it.

# Computer Interaction

CLI will respond to execution of commands by providing some output. Output should with the same reasons as input be consistent and predictable, even further it should follow some industry standards and thus allow great strength of CLI use: composability. Output should be served in the right dosage. We shouldn't output lines and lines of data because we will shaddow the important part the same way we shouldn't let the CLI command run in background not providing any output for minutes as it is working because the user will start to doubt if the command is broken. It's important to get this balance right.

## Interaction
Commands must be stateless and idempotent, meaning they can be run without relying on external state. If we have some external state like url config each command must include an option to define it on the fly, but it must also work without by first relaying on the externally saved state in config and if not found relaying on default value. If when relaying on default value the value creates an error there should be an explanation for the user that config must be created. We try to use default values first to get that “works like magic” feeling.

Never require a prompt for user input. Always offer a flag to provide that input, however if a user doesn't provide required input we can offer a prompt as an alterantive.


## Output

“Expect the output of every program to become the input to another, as yet unknown, program.” — Doug McIlroy

Output should use **stdout**. Output of commands should be presented in a clear formatted way for users to easily scan it, but it must also be compatible with `grep` command often used in command chaining. Output should also be possible in json by using `--json` flag. 

Default command response should be to the stdout and not saved to files. Anytime we want the output to be saved to a file we should explicitly specify so by using `--output filename.txt` flag and providing the path.


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
Error should be human readable, avoid printing out stack trace. It should give some pointers as to what could have gone wrong and how to fix it. Maybe try involving an example. However allow `--verbose` flag for complete error info with stack. Error messages should be sent to `stderr`

```
Connection to Flow Emulator was not successful, please make sure your api url is correct.

Set it by using: flow config 

Connection Error: connection error: desc = "transport: Error while dialing dial tcp [::1]:3000: connect: connection refused"
```

## Saving
Saving output to files should be optional and never a default action unless it is clear from the command name the output will be saved and that is required for later use (example: `flow init`). Output can be piped in files and our CLI should resort to that for default saving. If there is a specific format that needs to be saved or converted to we should be able to display that format as well hence again allowing us to pipe it to files.

## Help
Main help screen must contain a list of all commands with their own description, it must also include examples for commands. Help screen should include link to the documentation website.

Commands should have a description that is not too long (less than 100 characters). Help should be outputed if command is ran without any arguments. Help outline should be:

```Description:
    <description>

Usage:
    <command> <action>
    <command> <action>
    …

Examples:
    An optional section of example(s) on how to run the command.

Commands:
    <command>
        <commandDescription>
    <command>
        <commandDescription>
    …

Options:
    --<flag>
        <flagDescription>
    --<flag>
        <flagDescription>
    …
```

## Progress
Show progress for longer running commands visually. Visual feedback on longer running commands is important so users don’t get confused if command is running or the CLI is hang up resulting in user quitting. 

## Feedback
Commands should provide feedback if there is no result to be presented. If a command completes an operation that has no result to be shown it should write out that command was executed. This is meant to assure the user of the completion of the command.
If user is executing a destructive command they should be asked for approval but this command should allow passing confirmation with a flag --yes. 

## Exit
CLI should return zero status unless it is shut down because of an error. This will allow chaining of commands and interfacing with other cli. If a command is long running then provide description how to exit it (press ctrl+c to stop).
```
exit status 1
```

## Colors
Base color should be white with yellow reserved for warnings and red for errors. Blue color can be used to accent some commands but it shouldn’t be used too much as it will confuse the user.





# Inspiration and Reference 
https://clig.dev/

https://blog.developer.atlassian.com/10-design-principles-for-delightful-clis/

https://devcenter.heroku.com/articles/cli-style-guide

https://eng.localytics.com/exploring-cli-best-practices/








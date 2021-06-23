# Command Execution Rules

## _This pages describes functionality that is still in development._

## Rule Structure

Rules help Gort to determine who is able to perform what task. Gort rules follow a specific format. The rule structure describes what command is executed and what permission is needed in order to execute the command. If a user does not have the specified permission, the user is not able to execute the command.

The general form of a command is:

```
COMMAND [when CONDITIONS] [allow|must have PERMISSION]
```

* Command: The command indicates the command that's affected by the rule. Commands are referred to as `bundle_name:command_name`. For example, the `splitecho` command in the `echo` bundle would be referenced as `echo:splitecho`.

* Conditions: The (optional) conditions clause indicates when the rule should be is applied. It starts with the keyword `when`, and consists of one or more logical statements. See below for more detail. If a rule contains no conditions, it _always_ applies when the command is used.

* Permissions: The permissions clause indicates the permissions that a user must have to execute the command when the conditions are met. It begins with the phrase `must have`. Like commands, permissions are namespaced: `bundle_name:permission_name`.

* Allow: The standard permissions clause may be replaced with the `allow` keyword, which can be used to allow a command meeting the rule conditions to be executed by any Gort user. `allow` is used in lieu of a permissions clause, and may not be accompanied with any other keyword or phrase.

A basic example of a rule is:

```
foo:bar with option[delete] == true must have foo:destroy
```

This rule states that a user attempting to use the `bar` command from the `foo` bundle, with the `delete` flag set, must have the `foo:destroy` permission.

Rules can also be used to grant broad permissions by using the `allow` keyword:

```
foo:biz allow
```

This is the simplest possible rule, which allows any user to use the `foo:biz` command under all conditions.

## The Conditions Clause

The conditions rule clause begins with the keyword `with`.

The `conditions` clause can match specific command parameter, allowing you to create rules that apply under very specific invocations of a command.

### Options and Arguments

Any command can have two kinds of command parameters: _options_, are a general term for command flags and switches, and _arguments_, which are the main inputs into the command.

For example, given the following command:

```
curl -I --capath /home http://example.com
```

The options are `-I` and `--capath /home`, and the parameter is `http://example.com`

### Testing Options and Arguments

<!-- Thought: do we want to eventually add support for built-in functions in conditions, like time-based functions? Maybe we can allow inspection of the user's attributes? -->

Each rule can reference two pre-defined two data structures: `option` and `arg`.

* `option`: A map or dict of the commands options. The value of specific options can be accessed using standard map notation.

* `arg`: A (zero-indexed) list of the command arguments. Specific arguments can be accessed using standard map notation.

### Logical Operators

Individual (non-collection) values can logically evaluated using the `<`, `>`, `==` and `!=` operators:

* `with option["dry-run"] == true`

Regular expressions may also be used.

* `with option["set"] == /.*/`

Not only can specific `arg` positions be referenced by index, the entire parameter list can also be evaluated as a string by omitting the index. For example, given the following command:

```
echo foo bar
```

The following statements are equivalent:

* `foo:bar with arg[0] == 'foo' and arg[1] == 'bar' allow`
* `foo:bar with arg == 'foo bar' allow`

### Sets

Options and arguments can be tested against sets of conditions by using one of the following keywords:

* `in` -- Applied to a non-collection value, resolves to true if and only if the value matches a value in the set.
* `any`, `in` -- Applied to a collection value, resolves to true if and only if any value in the collection matches a value in the set.
* `all`, `in`-- Applied to a collection value, resolves to true if and only if all value in the collection match a value in the set.

Conditional sets can include zero or more values between square brackets. Regular expressions are also legal members and will be evaluated accordingly. Some examples are:

* `foo:bar with arg[0] in ['baz', false, 100] must have foo:read`
* `foo:bar with option["foo"] in ["foo", "bar"] allow`
* `foo:bar with any option == /^prod.*/ must have foo:read`
* `foo:bar with any arg in ['wubba'] must have foo:read`
* `foo:bar with any arg in ['wubba', /^f.*/, 10] must have foo:read`
* `foo:bar with all arg in [10, 'baz', 'wubba'] must have foo:read`
* `foo:bar with all option < 10 must have foo:read`
* `foo:bar with all options in ['staging', 'list'] must have foo:read`

### Combining Qualifiers

Arbitrarily long compound qualifiers can be constructed using the `and` and/or `or` keywords, so your rules can be as simple or as complicated as you need them to be. For example, the following rule is legal:

```
foo:bar with arg=="prod" and option["delete"] == true or option["set"] == /.*/ must have foo:destroy
```

## Permissions

The permissions clause is where you state any permissions that are required to execute the command. The beginning of the permissions clause is indicated by the phrase `must have`.

Like the conditions clause, it can be arbitrarily complex, and can a single permission, a specific combination of permissions combination, or a list of permissions. It supports the same operations as well:

* `or`
* `and`
* `any in`
* `all in`
* `allow`

For example, the following are rule examples with valid permission settings:

* `foo:baz with option[delete] == true must have foo:write and site:admin`
* `foo:export must have all in [foo:write, site:ops] or any in [site:admin, site:management]`
* `foo:bar must have any in [foo:read, foo:write]`
* `foo:qux must have all in [foo:write, site:ops] and any in [site:admin, site:management]`
* `foo:biz allow`

Note the special `allow` keyword, which can be used in lieu of a permissions clause to allow a command to be executed by any registered user in Gort.

## Todo

1. Built-in/standard permissions (esp. for Gort administration actions)
1. Built-in functions in conditions?
1. Access to user/group/adapter attributes in conditions?

<!-- ## Site Namespace

The site namespace is used when trying to set permissions for a user, group, or role. This does not have to be command specific. You may use site permissions when deciding what group should have permissions to execute certain commands, in specific environments, within certain groups.

A user can only create and delete permissions from the site namespace. You cannot delete the permissions that are part of a command bundle.

For example, let's say your organization has an IT group, "it", an engineering group, "eng", and a QA group, "qa". As a result, you have 3 different environments "prod", "test" and "stage". There are certain tasks that can be performed in each environment, but you must belong to the correct group and be operating in the correct environment.

So we will assume that The IT group operates in "prod", QA in "qa", and Engineering in "staging", though IT should be able to handle certain tasks in all environments such as patch updates and the sort.

Let's create some example commands: `foo:deploy`, `foo:patch`, `foo:delete`, `foo:readlog`

For the examples sake, we'll have the example permissions map to these commands such that they may look like: `foo:p_deploy`, `foo:p_patch`, `foo:p_delete`, `foo:p_readlog`

We'll set up site permissions based on each group and each environment: `site:prod`, `site:test`, `site:stage`, `site:it`, `site:qa`, `site:eng`.

Some resulting rules may look like the following:

* `foo:deploy with option[environment] == 'prod' must have all in [site:it, site:prod, foo:p_deploy]`
* `foo:deploy with option[environment] == 'qa' must have site:test and foo:p_deploy`
* `foo:deploy with option[environment] == 'stage' must have site:stage and foo:p_deploy`
* `foo:patch must have all in [foo:p_patch, site:it] or all in [site:qa, site:test, foo:p_patch] or all in [site:eng, site:stage, foo:p_patch]` -->

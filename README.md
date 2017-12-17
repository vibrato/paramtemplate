ParamTemplate
=============

ParamTemplate is a small application that allows you to do template driven
connections to AWS (and other) secret stores to generate environment specific
config files.

In researching how to use AWS's ParamStore to actually fill config files all
documentation points to calling the CLI and exporting it as either an
environment variable or replacing it into a file. I particularly agree with the
environment variable method however this is not always an option. When you are
managing many secrets and config options this becomes confusing.

ParamTemplate let's you define the secrets you need directly in the template.

Usage
=====

ParamTemplate is base on Go's template functionality and so in general you
should write go templates.

On top of this we have implemented a number of functions like `ssmGet` or
`ssmGetPath` that indicate we should fetch a parameter from `ssm`.

A specific example (and a key reason for writing this) is dealing with
SetParameter.xml files for IIS and webdeploy. Secrets are embedded in the xml
to be replaced at installation time.

Example::

```xml
<parameters>
  <setParameter name="IIS Web Application Name"
                value="My Web Site" />
  <setParameter name="Identity-Web.config Connection String"
                value="{{ ssmGet "/iis/Identity-Web.config_Connection_String" "decrypt=true" }}" />
  <setParameter name="Shopping-Web.config Connection String"
                value="{{ ssmGet "/iis/Shoppipng-Web.config_Connection_String" "decrypt=true" }}"/>
</parameters>
```

We can then replace execute this template like:

```
export AWS_PROFILE=myaccount
export AWS_REGION=ap-southeast-2

./paramtemplate < mytmpl.gotmpl > output.xml
```

In this example we read from stdin and write to stdout, but you can also use
the `-template` and `-output` flags to specify file paths.

From a higher level usage I bake the application into a cloud image and then
invoke it as part of the userdata script of an autoscaling group.

Authentication
==============

Autentication to the cloud provider is to be handled entirely outside of the
ParamTemplate application. By using the AWS Go SDK ParamTemplate will
automatically respect the `AWS_PROFILE` and `AWS_SECRET_ACCESS_KEY` style
environment variables. You should set these variables, or configure Instance
Profiles externally to the application.

Function Parameters
===================

A side effect of golang and go templates is that you can't directly pass
optional parameters to functions. To get around this functions may take
parameters in the form `key=value` which will be transformed into an optional
parameter to the function.

From the above example::

```
{{ ssmGet "/iis/Identity-Web.config_Connection_String" "decrypt=true" }}
```

The `ssmGet` function takes a boolean parameter for whether it should try and
decrypt the string from the AWS key store. By default this is false (because
it's false in all the APIs, IMO should default true), but you can optionally
pass the parameter to invoke decryption.

Different functions will support different parameters.

Future Work
===========

The most obvious future work is to implement other provider key stores. There
is nothing about the core of this that is specific to AWS and it should be
fairly easily extended to Azure, GCE and Hashicorp Vault.

Supported Functions
===================

ssmGet
------

Fetches a single secret from AWS ParamStore

*Parameters*

* `decrypt`:`bool` - Decrypt the secret when retrieving

*Return*

Returns a Parameter object which can either be converted directly to a string like:

```
{{ ssmGet "/iis/Identity-Web.config_Connection_String" "decrypt=true" }}
```

or as an Object with `Name`, `Version`, `Value` Properties so you can do things like:

``
{{ with ssmGet "/ecomm/Identity-Web.config_Connection_String" "decrypt=true" }}
<setParameter name="Identity-Web.config Connection String" value="{{ .Value }}" /><!-- Version: {{ .Version }} -->
{{ end }}
``

ssmGetPath
----------

Fetches a group of secrets by Path from AWS ParamStore:

*Parameters*

* `decrypt`:`bool` - Decrypt the secrets when retrieving
* `maxresults`:`int` - Return a maximum number of secrets
* `recursive`:`bool` - Recurse into subpaths to fetch secrets
* `trim`:`bool` - Remove the provided path name from the start of the secret name.

*Returns*

Returns an array of Parameters similar to `ssmGet`

*Example*

```
---
ThingsInYaml:
  {{ range ssmGetPath "/my-test/" "decrypt=true" "trim=true" -}}
  - {{ .Name }}: {{ .Value }}   # {{ .Version }}
  {{ end }}
```

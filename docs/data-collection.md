---
title: Data Collection
description: Data collected from Flow CLI usage
---

Flow CLI tracks flow command usage count using Mixpanel.

Data collection is enabled by default. Users can opt out of our data collection through running `flow settings metrics disable`. 
To opt back in, users can run `flow settings metrics enable`.

## Why do we collect data about flow cli usage?

Collecting aggregate command count allow us to prioritise features and fixes based on how users use flow cli.

## What data do we collect?

We only collect the number of times a command is executed. 

We don't keep track of the values of arguments, flags used
and the values of the flags used. We also don't associate any commands to any particular user.

The only property that we collect from our users are their preferences for opting in / out of data collection. 
The analytics user ID is specific to Mixpanel and does not permit Flow CLI maintainers to e.g. track you across websites you visit.

Further details regarding the data collected can be found under Mixpanel's data collection page in `Ingestion API` 
section of https://help.mixpanel.com/hc/en-us/articles/115004613766-Default-Properties-Collected-by-Mixpanel.

Please note that although Mixpanel's page above mentions that geolocation properties are recorded by default, 
we have turned off geolocation data reporting to Mixpanel.

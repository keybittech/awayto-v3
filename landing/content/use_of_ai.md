---
title: "Use of Artificial Intelligence and Large-Language Models"
weight: 1
---

## Guiding Principle

Large-language models (LLMs) will only be used as a supplement to basic functionality on the site. There is no intention to use LLMs to automate user workflows, intercept business data for advanced use cases, or provide direct access to a chatbot or other similar technology.

## Opt-In, Off by Default

An option is provided during group creation to enable AI suggestions across the group. This can be toggled on or off at any time later on, using the group configuration screens.

## How is AI used on the website?

At this time, the OpenAI API is used to facilitate LLM functionality as a means of bootstrapping faster development. In the future, the platform will utilize local language models, namely Ollama, a leading free and open-source local LLM framework.

## What basic features are supplemented by AI?

AI is used to provide suggestions pertaining to the setup and configuration of a Group along with its supporting objects (Roles, Services, Tiers, and Features).

## Group Naming Suggestions

Some groups or organizations which use the site may find it difficult to enumerate the concepts and naming structures behind their services, when it comes to declaring them in a schedulable format that users will understand. Therefore, AI is used to help supplement idea generation when it comes to creating a group. 

Upon group creation, the group owner must provide a group name and short phrase of the group's purpose. **Only  group name & group purpose -- along with the names of subsequently-created Roles, Services, Tiers, or Features -- are used to query for AI generated suggestions.**

### What it looks like

When creating group objects (Roles, Services, Tiers, and Features) the user will be made aware that AI is in use by the following method:

1. Before a supported form element has fetched AI suggestions, the form element will show generic examples, denoted with "Ex:"

![an html dropdown input with example options listed under; the options are not clickable](/doc_images/pre_ai.png)

2. After suggestions have been fetched, they will be shown under the form element, denoted with "AI:"

![an html dropdown input with ai-generated options listed under; each option is clickable text](/doc_images/post_ai.png)

3. The user may click on a suggestion if desired, and it will be added to the field.

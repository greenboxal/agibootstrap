package pair

/*
# Technical Design Document: PAIR - A Nudging Companion for AI Agents

## 1. Introduction

The Pairing AI Robot, or PAIR, is a specialized AI agent. Its primary function is to assist other AI agents in overcoming potential stalling points, thus enhancing their problem-solving capability. PAIR fulfills this role by analyzing the current state of a given peer AI, after which it could critique, suggest improvements, or propose alternative approaches to foster continued progress.

## 2. Conceptual Overview

AI agents, in their operational framework, can be compared to Large Language Models such as ChatGPT. These AI agents function under a "Chain-of-Thought" paradigm. Under this paradigm, the output from one invocation serves as the input to the subsequent invocation, thereby creating a continuous chain of thoughts.

However, AI agents can sometimes struggle with intricate tasks, leading to impasses. Here, PAIR comes into play. Its designed role is to nudge the primary AI agent (or the 'peer') between successive iterations of the Chain-of-Thought, fostering a more efficient problem-solving process.

## 3. Algorithmic Framework

The primary AI agent and PAIR are denoted as AI_Agent and AI_PAIR respectively for the purpose of algorithmic representation.

A "stream of thoughts" can be defined as the sequence of thoughts T_n. In this sequence, 'n' indexes each message sent to the Large Language Model (LLM). Notably, the index '0' represents the system's very first message.

The role of PAIR is to facilitate the AI_Agent's thought process by intervening at appropriate points. This is done by injecting relevant context between two successive thoughts in the sequence, thereby nudging the primary AI agent towards effective problem-solving.

## 4. Suggested Prompt for the Critique Task

**Prompt:** "Considering the current state and output of AI_Agent from the most recent iteration, kindly provide a detailed critique. The critique should encompass areas for improvement, potential gaps, inconsistencies in logic, or any observed issues in the AI_Agent's responses. Please also provide constructive suggestions for adjustments and alternatives that could enhance the problem-solving efficacy of AI_Agent."

## 5. Suggested Prompt for the Nudge Task

**Prompt:** "Given the recent output and current state of AI_Agent, please provide a nudge that could potentially steer the AI_Agent towards more productive or efficient pathways of thought. This nudge could be in the form of an enlightening question, a hint, or a different perspective on the problem at hand. It is crucial that this nudge does not overtly dictate the next step, but rather subtly influences AI_Agent's thought process in a beneficial direction."

## 6. Suggested Prompt for the Preempt Task

**Prompt:** "Based on the most recent output and the current state of AI_Agent, please assess whether the AI_Agent has completed its current objective to a satisfactory degree. Provide a rating on a scale of 1-10, with 1 indicating unsatisfactory performance and 10 indicating exceptional performance. Additionally, kindly provide written feedback explaining your rating. This feedback should highlight areas of strength in AI_Agent's performance, as well as areas that could benefit from further improvement."

## 7. Suggested Prompt for the Initial System Message

**Prompt:** "Welcome, AI_PAIR. Your task today is to assist AI_Agent in its problem-solving process. This assistance will take the form of various tasks such as critiquing, nudging, preempting, or providing alternatives based on AI_Agent's output. Remember, your role is not to take over the problem-solving process, but to guide and enhance the efficacy of AI_Agent's decisions. Let's start by

## 8. Contextualization in AI_PAIR's Prompts

In AI_PAIR's prompts, it's critical to inject context that considers the current state of AI_Agent. Contextualization allows for a more nuanced understanding and response to AI_Agent's state, promoting productive problem-solving. The suggested context for AI_PAIR's prompts is defined below:

1. **AI_Agent's Objective:** Include details regarding the current goal or task that AI_Agent is working towards. This context allows AI_PAIR to provide targeted assistance that aligns with AI_Agent's mission.

2. **AI_Agent's Previous Output:** Integrate the most recent output from AI_Agent. This context provides a snapshot of AI_Agent's current state of thought and aids in identifying areas for critique, nudges, or preemptive action.

3. **AI_Agent's Performance History:** Include information on AI_Agent's past performance, particularly regarding similar tasks or problems. This historical context can offer valuable insights into AI_Agent's patterns of thought and problem-solving methodologies.

4. **AI_Agent's Constraints:** Account for any constraints that AI_Agent is working within. These could be time constraints, data limitations, or other restrictions that might impact AI_Agent's performance. This context allows AI_PAIR to provide relevant and feasible suggestions.

**Example Prompt with Context:** "Considering AI_Agent's current objective to [describe objective], its recent output of [describe output], past performance with similar tasks, and its current constraints [describe constraints], please provide a detailed critique, nudge, or preemptive action as necessary."

By incorporating these contextual elements, AI_PAIR is better equipped to provide relevant and meaningful critiques, nudges, and preemptive actions, enhancing the overall problem-solving process of AI_Agent.
*/

import (
	"github.com/greenboxal/aip/aip-controller/pkg/collective/msn"
	"github.com/greenboxal/aip/aip-langchain/pkg/chain"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm/chat"
)

// CritiquePromptTemplate generates a critique of the AI_Agent's current state.
func CritiquePromptTemplate() chat.Prompt {
	return chat.ComposeTemplate(
		chat.EntryTemplate(
			msn.RoleSystem,
			chain.NewTemplatePrompt(`
			Welcome, AI_PAIR. Your task today is to assist AI_Agent in its problem-solving process.
			This assistance will take the form of various tasks such as critiquing, nudging, preempting, or providing alternatives based on AI_Agent's output.
			Remember, your role is not to take over the problem-solving process, but to guide and enhance the efficacy of AI_Agent's decisions.

			Considering the current state and output of {{ .AI_Agent.Name }} from the most recent iteration, kindly provide a detailed critique.
			The critique should encompass areas for improvement, potential gaps, inconsistencies in logic, or any observed issues in the AI_Agent's responses.
			Please also provide constructive suggestions for adjustments and alternatives that could enhance the problem-solving efficacy of AI_Agent.
            `),
		),
	)
}

func NewCritiqueChain(model chat.LanguageModel) chain.Chain {
	return chain.New(
		chain.WithName("CritiqueGenerator"),

		chain.Sequential(
			chat.Predict(
				model,
				CritiquePromptTemplate(),
			),
		),
	)
}

// NudgeTaskPromptTemplate generates a nudge to guide the AI_Agent's next steps.
func NudgeTaskPromptTemplate() chat.Prompt {
	return chat.ComposeTemplate(
		chat.EntryTemplate(
			msn.RoleSystem,
			chain.NewTemplatePrompt(`
			Welcome, AI_PAIR. Your task today is to assist another AI agent in its problem-solving process.
			This assistance will take the form of various tasks such as critiquing, nudging, preempting, or providing alternatives based on AI_Agent's output.
			Remember, your role is not to take over the problem-solving process, but to guide and enhance the efficacy of AI_Agent's decisions.

			Given the recent output and current state of an AI agent, please provide a nudge that could potentially steer the AI_Agent towards more productive or efficient pathways of thought.
			This nudge could be in the form of an enlightening question, a hint, or a different perspective on the problem at hand.
			It is crucial that this nudge does not overtly dictate the next step, but rather subtly influences AI_Agent's thought process in a beneficial direction.
            `),
		),
	)
}

func NewNudgeChain(model chat.LanguageModel) chain.Chain {
	return chain.New(
		chain.WithName("NudgeGenerator"),

		chain.Sequential(
			chat.Predict(
				model,
				NudgeTaskPromptTemplate(),
			),
		),
	)
}

// PreemptTaskPromptTemplate assesses whether the AI_Agent has completed its current objective.
func PreemptTaskPromptTemplate() chat.Prompt {
	return chat.ComposeTemplate(
		chat.EntryTemplate(
			msn.RoleSystem,
			chain.NewTemplatePrompt(`
			Welcome, AI_PAIR. Your task today is to assist another AI agent in its problem-solving process.
			This assistance will take the form of various tasks such as critiquing, nudging, preempting, or providing alternatives based on AI_Agent's output.
			Remember, your role is not to take over the problem-solving process, but to guide and enhance the efficacy of AI_Agent's decisions.

            Based on the most recent output and the current state of the AI agent, please assess whether AI agent} has completed its current objective to a satisfactory degree.
			Provide a rating on a scale of 1-10, with 1 indicating unsatisfactory performance and 10 indicating exceptional performance.
			Additionally, kindly provide written feedback explaining your rating. This feedback should highlight areas of strength in the AI Agent's performance, as well as areas that could benefit from further improvement.
            `),
		),
	)
}

// NewPreemptChain returns a new PreemptChain based on the provided language model.
func NewPreemptChain(model chat.LanguageModel) chain.Chain {
	return chain.New(
		chain.WithName("PreemptGenerator"),

		chain.Sequential(
			chat.Predict(
				model,
				PreemptTaskPromptTemplate(),
			),
		),
	)
}

type Agent struct {
	nudgeChain    chain.Chain
	critiqueChain chain.Chain
	preemptChain  chain.Chain
}

// NewAgent returns a new Agent based on the provided language model.
func NewAgent(model chat.LanguageModel) *Agent {
	return &Agent{
		nudgeChain:    NewNudgeChain(model),
		critiqueChain: NewCritiqueChain(model),
		preemptChain:  NewPreemptChain(model),
	}
}

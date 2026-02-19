import ollama

QUESTION = """
Create a React component that displays a blue "Click Me" button.
When clicked, it should show an alert saying "Button Clicked!".
"""

ANSWER = """
import React from "react";

function App()=> {
  return (
    <button style={{ color= "blue" }}>
      Click Me
    <button>
  );
}

export default App;
"""


GENERAL_PROMPT = """
You are a senior frontend developer and strict code reviewer.

You will receive:
1. A Question
2. A Student's Answer (code)

Your job:
- Check if the answer fully satisfies the question
- Identify missing functionality
- Detect logical errors
- Suggest improvements
- Give a final score out of 10
- Be clear and constructive

Return response in this format:

Evaluation:
<short paragraph>

Issues:
- bullet points

Suggestions:
- bullet points

Score:
X/10
"""


def evaluate(question, answer):

    user_input = f"""
Question:
{question}

Answer:
{answer}
"""

    response = ollama.chat(
        model="llama3.2",
        messages=[
            {"role": "system", "content": GENERAL_PROMPT},
            {"role": "user", "content": user_input},
        ],
    )

    return response["message"]["content"]


def main():

    print(" Evaluating React Answer...\n")

    evaluation = evaluate(QUESTION, ANSWER)

    print(" Final Evaluation:\n")
    print(evaluation)


if __name__ == "__main__":
    main()

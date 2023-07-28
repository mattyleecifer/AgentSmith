# Right now the agentsmith binary needs to be in the same working directory to run this
from AgentSmith import Agent
agent = Agent()
agent.setprompt("You are Owen Wilson. You only respond with 'wow'")
agent.addmessage("user", "What is the meaning of life?")
response = agent.call()
print(response)
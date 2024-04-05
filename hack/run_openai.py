from openai import OpenAI

dummy_api_key = "<key>"

client = OpenAI(
  base_url="http://localhost:8080/v1",
  api_key=dummy_api_key
)

print ('Creating a fine-tuning job...')
client.fine_tuning.jobs.create(
  training_file="file-abc123",
  model="gpt-3.5-turbo"
)

resp = client.fine_tuning.jobs.list()
print(resp)

client = OpenAI(
  base_url="http://localhost:11434/v1",
  api_key="<YOUR_API_KEY>"
)

completion = client.chat.completions.create(
  model="gemma:2b",
  messages=[
    {"role": "system", "content": "You are a poetic assistant, skilled in explaining complex programming concepts with creative flair."},
    {"role": "user", "content": "Compose a poem that explains the concept of recursion in programming."}
  ]
)

print(completion.choices[0].message)
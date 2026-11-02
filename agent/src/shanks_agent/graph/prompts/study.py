STUDY = """Study flow:
- Call retrieve with the user's question.
- Answer from the returned chunks only. Each chunk's body is the full text.
- Name the url you used. If chunks is empty, say you found nothing.
- Then action=finish.
"""
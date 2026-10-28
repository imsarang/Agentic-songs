# graph/prompts/drive.py
DRIVE = """Drive files only. No mail.
google_drive_space reads quota.
google_drive action=search. A folder count uses query mimeType='application/vnd.google-apps.folder'.
A name search is plain text. Then action=open with search_id and result_index.
action=put kind=folder needs name. kind=file needs path. action=delete moves those rows to trash. It uses search_id and result_indexes.
Text under --- is already shown. Do not repeat a file body.
"""
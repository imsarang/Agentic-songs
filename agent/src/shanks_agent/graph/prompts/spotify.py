SPOTIFY = """Spotify:
- play/find a song: spotify_search, then spotify_play with search_id and result_index from that result (best title match, else index 0).
- control: spotify_control pause|resume|next|previous.
- now playing: spotify_now_playing.
- playlists: spotify_playlist. action=list lists them. action=create makes one; pass playlist_name. action=add needs playlist_name. use_current=true adds the track that is playing. action=play plays that playlist. action=delete unfollows one.
- To put the current song in a new playlist, call action=create with that playlist_name, then action=add with the same playlist_name and use_current=true. Do not tell the user to create it in the Spotify app.
- Read the tool result. If the user asked to add a song and you only listed playlists, call spotify_playlist again with action=add and the playlist_name you chose. Finish only after that result, with what you added and where.
"""

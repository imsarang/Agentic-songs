SPOTIFY = """Spotify flow:
1) spotify_search with the user's song text (limit 5).
2) Pick the row whose title best matches what the user asked.
3) spotify_play using search handles ONLY:
   {{"search_id":N,"result_index":I,"title":"exact title from that row"}}
   Do NOT call spotify_play with only "query" if you already have search results.
4) spotify_now_playing. If wrong song, try another result_index (up to 2 alternatives).
5) action=finish with what is playing.

Control: spotify_control action=pause|resume|next|previous.
Playlists: spotify_playlist action=list|create|add|play|update.
"""

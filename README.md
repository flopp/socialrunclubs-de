# socialrunclubs-de
A directory of "Social Run Clubs" in Germany https://socialrunclubs.de

# Features
* displayed club data:
  * description
  * links to Instagram, Strava, WhatsApp, TikTok, club website
  * map of meeting point (if known)
  * link to Google Maps
* run clubs by city
* overview maps showing all run clubs
* club + city search

# Design:
* mobile first & clean
* based on Blades CSS (https://blades.dev/)

# Implementation
* data is stored in Google Sheets
* go based static site generator (pulls data from Google Sheets and produces static HTML files)

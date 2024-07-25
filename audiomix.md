# Audio Mixing

To keep the audio decent, follow these rules:

## General Rules

- Announcer / Speaker
  - -12 dB to -6 dB
  - prefer slightly dipping into the red range instead of lowering volume
  - keep background noise below -20 dB **at this source** (this will come later)
  - consider using light compression
- Sound Effects
  - peak around -15 dB to -10 dB
  - keep balance with announcer / speaker
- Background Noise
  - have a mic close to where crowds of people are
  - more = better (pretty much)
  - peak around -20 dB to -15 dB
  - use a noise gate or something similar to prevent sudden spikes
  - consider raising volume after a big event (e.g., big win, award)

## Implementing in OBS

1. **setting up audio sources:**
   - add your audio sources (e.g., microphone, sound effects, background noise) in OBS
   - label each source clearly for easy identification

2. **adjusting audio levels:**
   - click the gear icon next to each audio source in the audio mixer panel
   - go to "filters" and add a "gain" filter to each source
     - for the announcer/speaker, set the gain to ensure levels are between -12 dB and -6 dB
     - for sound effects, set the gain to ensure levels peak around -15 dB to -10 dB
     - for background noise, set the gain to ensure levels peak around -20 dB to -15 dB

3. **applying compression:**
   - for the announcer/speaker, add a "compressor" filter
     - set the threshold just below your desired peak level (e.g., -12 dB)
     - adjust the ratio to 3:1 or 4:1 for light compression
     - fine-tune the attack and release settings to ensure natural sound

4. **using noise gates:**
   - for background noise, add a "noise gate" filter
     - set the close threshold to just below the noise level you want to cut out
     - set the open threshold slightly above the noise level to allow desired audio through
     - adjust the attack, hold, and release settings to ensure smooth transitions

5. **balancing audio:**
   - monitor the overall audio mix in the audio mixer panel
   - make real-time adjustments to ensure the announcer/speaker, sound effects, and background noise are balanced
   - ensure the background noise is prominent but does not overpower the announcer/speaker

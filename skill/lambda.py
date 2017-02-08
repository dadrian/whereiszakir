#!/usr/bin/python
# coding=utf8
import json
import decimal
import datetime
import requests


# --------------- Helpers that build all of the responses ----------------------
 
def build_speechlet_response(title, output, reprompt_text, should_end_session):
    return {
        'outputSpeech': {
            'type': 'PlainText',
            'text': output
        },
        'card': {
            'type': 'Simple',
            'title': title,
            'content': output
        },
        'reprompt': {
            'outputSpeech': {
                'type': 'PlainText',
                'text': reprompt_text
            }
        },
        'shouldEndSession': should_end_session
    }


def build_response(session_attributes, speechlet_response):
    return {
        'version': '1.0',
        'sessionAttributes': session_attributes,
        'response': speechlet_response
    }


# --------------- Functions that control the skill's behavior ------------------

def get_location():
    r = requests.get('https://whereiszakir.com/where')
    if r.status_code != 200 or 'location' not in r.json():
        return u"¯\_(ツ)_/¯"
    else:
        return r.json()['location']

def get_welcome_response(session_attributes):
    """ If we wanted to initialize the session to have some attributes we could
    add those here
    """
    location = get_location()
    card_title = location
    speech_output = ""
    if location == u"¯\_(ツ)_/¯":
        speech_output = "I don't know where Zakir is today, sorry."
    else:
        speech_output = "Today, Zakir is in "+location
    reprompt_text = ""
    should_end_session = True
    return build_response(session_attributes, build_speechlet_response(
        card_title, speech_output, reprompt_text, should_end_session))


def handle_session_end_request():
    card_title = ""
    speech_output =""
    should_end_session = True
    return build_response({}, build_speechlet_response(
        card_title, speech_output, None, should_end_session))

# --------------- Events ------------------

def on_launch(launch_request, session):
    """ Called when the user launches the skill without specifying what they
    want
    """
    # Dispatch to your skill's launch
    return get_welcome_response(session['attributes'])

def on_intent(intent_request, session):
    """ Called when the user specifies an intent for this skill """
    return get_welcome_response(session['attributes'])

def on_session_ended(session_ended_request, session):
    """ Called when the user ends the session.
    Is not called when the skill returns should_end_session=true
    """
    # add cleanup logic here
    return handle_session_end_request()

# --------------- Main handler ------------------

def lambda_handler(event, context):
    """ Route the incoming request based on type (LaunchRequest, IntentRequest,
    etc.) The JSON body of the request is provided in the event parameter.
    """
    if (event['session']['application']['applicationId'] != "amzn1.ask.skill.ad8bbc74-7271-4439-ac98-78ce67602315"):
         raise ValueError("Invalid Application ID")

    if event['request']['type'] == "LaunchRequest":
        return on_launch(event['request'], event['session'])
    elif event['request']['type'] == "IntentRequest":
        return on_intent(event['request'], event['session'])
    elif event['request']['type'] == "SessionEndedRequest":
        return on_session_ended(event['request'], event['session'])
    return "{}"

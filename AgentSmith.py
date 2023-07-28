import subprocess
import os

class Agent:
    def __init__(self):
        self.key = None
        self.homeDir = None
        self.savechatName = None
        self.loadfileName = None
        self.prompt = None
        self.model = None
        self.maxtokens = None
        self.functions = []
        self.messages = []
        self.autofunction = None
        self.autoclearfunctionoff = None
        self.autorequestfunction = None

        self.callargs = []
    
    def setkey(self, new_key):
        self.key = new_key

    def sethomeDir(self, dir):
        self.homeDir = dir

    def save(self, filename):
        self.savechatName = filename
    
    def load(self, filename):
        self.loadfileName = filename

    def setprompt(self, new_prompt):
        self.prompt = new_prompt
    
    def setmodel(self, new_model):
        self.model = new_model

    def setmaxtokens(self, new_maxtokens):
        self.maxtokens = new_maxtokens
    
    def addfunction(self, new_function):
        self.functions.append(new_function)
    
    def addmessage(self, user, new_message):
        self.messages.append([user, new_message])
    
    def setautofunction(self):
        self.autofunction = True    
    
    def setautoclearfunctionfoff(self):
        self.autoclearfunctionoff = True

    def setautorequestfunction(self):
        self.autorequestfunction = True

    def createcall(self):
        self.callargs = []
        if self.key:
            self.callargs.append("-key")
            self.callargs.append(self.key)
        if self.homeDir:
            self.callargs.append("-home")
            self.callargs.append(self.homeDir)
        if self.savechatName:
            self.callargs.append("-save")
            self.callargs.append(self.savechatName)
        if self.loadfileName:
            self.callargs.append("-load")
            self.callargs.append(self.loadfileName)
        if self.prompt:
            self.callargs.append("-prompt")
            self.callargs.append(self.prompt)
        if self.model:
            self.callargs.append("-model")
            self.callargs.append(self.model)
        if self.maxtokens:
            self.callargs.append("-maxtokens")
            self.callargs.append(self.maxtokens)
        for i in self.functions:
            self.callargs.append("-function")
            self.callargs.append(i)
        for i in self.messages:
            if i[0] == "user":
                self.callargs.append("-message")
                self.callargs.append(i[1])
            if i[0] == "assistant":
                self.callargs.append("-messageassistant")
                self.callargs.append(i[1])
            if i[0] == "function":
                self.callargs.append("-messagefunction")
                self.callargs.append(i[1])
        if self.autofunction:
            self.callargs.append("-autofunction")
        if self.autoclearfunctionoff:
            self.callargs.append("-autoclearfunctionoff")
        if self.autorequestfunction:
            self.callargs.append("-autorequestfunction")
        
    def call(self):
        self.createcall()
        callargs = ' '.join(['{} "{}"'.format(arg, val.replace('"', '\\"')) if ' ' in val else '{} {}'.format(arg, val) for arg, val in zip(self.callargs[::2], self.callargs[1::2])])
        output = subprocess.check_output(os.getcwd() + "/agentsmith " + callargs, shell=True, text=True)
        return output
        
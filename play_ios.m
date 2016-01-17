// +build ios

// Framework includes
#import <AVFoundation/AVAudioSession.h>
#import <AudioToolbox/AudioToolbox.h>
#import <AVFoundation/AVFoundation.h>

#include "_cgo_export.h"

static OSStatus	render(
	void*                       inRefCon,
	AudioUnitRenderActionFlags* ioActionFlags,
	const AudioTimeStamp*       inTimeStamp,
	UInt32                      inBusNumber,
	UInt32                      inNumberFrames,
	AudioBufferList*            ioData
) {
	streamCallback(ioData->mBuffers[0].mData, inNumberFrames);
	for (UInt32 i = 1; i < ioData->mNumberBuffers; ++i) {
		memcpy(ioData->mBuffers[i].mData, ioData->mBuffers[0].mData, ioData->mBuffers[i].mDataByteSize);
	}
	return noErr;
}

AudioUnit au;

const char* nserrstr(NSString* msg, NSError* err) {
	return [[msg stringByAppendingFormat:@": %@", err.localizedDescription] cStringUsingEncoding:[NSString defaultCStringEncoding]];
}

const char* errstr(NSString* msg, OSStatus err) {
	return [[msg stringByAppendingFormat:@": %d", (int)err] cStringUsingEncoding:[NSString defaultCStringEncoding]];
}

const char* start() {
	AVAudioSession* s = [AVAudioSession sharedInstance];
	NSError* nserr = nil;

	[s setCategory:AVAudioSessionCategoryPlayback error:&nserr];
	if (nserr) return nserrstr(@"setCategory", nserr);

	NSNotificationCenter* nc = [NSNotificationCenter defaultCenter];
	[nc addObserverForName:AVAudioSessionInterruptionNotification object:s queue:nil usingBlock:^(NSNotification* n) {
		if ([[n.userInfo valueForKey:AVAudioSessionInterruptionTypeKey] intValue] == AVAudioSessionInterruptionTypeBegan) {
			OSStatus err = AudioOutputUnitStop(au);
			if (err) NSLog(@"AudioOutputUnitStop: %d", (int)err);
		} else { // AVAudioSessionInterruptionTypeEnded
			NSError* nserr = nil;
			[s setActive:YES error:&nserr];
			if (nserr) NSLog(@"AVAudioSession setActive: %@", nserr);

			OSStatus err = AudioOutputUnitStart(au);
			if (err) NSLog(@"AudioOutputUnitStart: %d", (int)err);
		}
	}];
	[nc addObserverForName:AVAudioSessionMediaServicesWereResetNotification object:s queue:nil usingBlock:^(NSNotification* n) { start(); }];

	[s setActive:YES error:&nserr];
	if (nserr) return nserrstr(@"setActive", nserr);

	AudioComponentDescription desc = {
		.componentType         = kAudioUnitType_Output,
		.componentSubType      = kAudioUnitSubType_RemoteIO,
		.componentManufacturer = kAudioUnitManufacturer_Apple,
	};
	OSStatus err = AudioComponentInstanceNew(AudioComponentFindNext(NULL, &desc), &au);
	if (err) return errstr(@"AudioComponentInstanceNew", err);

	const AudioUnitElement kOutputBus = 0;
	const AudioUnitElement kInputBus  = 1;

	AudioStreamBasicDescription fmt = {
		.mSampleRate       = 44100,
		.mFormatID         = kAudioFormatLinearPCM,
		.mFormatFlags      = kAudioFormatFlagIsFloat | kAudioFormatFlagsNativeEndian | kAudioFormatFlagIsPacked | kAudioFormatFlagIsNonInterleaved,
		.mBytesPerPacket   = 4,
		.mFramesPerPacket  = 1,
		.mBytesPerFrame    = 4,
		.mChannelsPerFrame = 1,
		.mBitsPerChannel   = 32,
	};
	err = AudioUnitSetProperty(au, kAudioUnitProperty_StreamFormat, kAudioUnitScope_Output, kInputBus, &fmt, sizeof(fmt));
	if (err) return errstr(@"StreamFormat", err);

	UInt32 one = 1;
	err = AudioUnitSetProperty(au, kAudioOutputUnitProperty_EnableIO, kAudioUnitScope_Output, kOutputBus, &one, sizeof(one));
	if (err) return errstr(@"EnableIO", err);

	AURenderCallbackStruct rc = {.inputProc = render};
	err = AudioUnitSetProperty(au, kAudioUnitProperty_SetRenderCallback, kAudioUnitScope_Output, kOutputBus, &rc, sizeof(rc));
	if (err) return errstr(@"SetRenderCallback", err);

	err = AudioUnitInitialize(au);
	if (err) return errstr(@"AudioUnitInitialize", err);

	err = AudioOutputUnitStart(au);
	if (err) return errstr(@"AudioOutputUnitStart", err);

	return nil;
}

const char* stop() {
	OSStatus err = AudioOutputUnitStop(au);
	if (err) return errstr(@"AudioOutputUnitStop", err);
	return nil;
}

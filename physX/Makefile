SOURCES = *.cpp
LIBRARY = CphysX.so
INSTALLDIR = ../../../../../bin

SDK=/usr/local/PhysX-3.2.1-SDK/Include
INC=-I$(SDK) -I$(SDK)/RepX -I$(SDK)/RepXUpgrader
CFLAGS  = -DPHYSX_PROFILE_SDK -D_DEBUG -DPX_CHECKED -DPX_SUPPORT_VISUAL_DEBUGGER $(INC) -Werror -m32 -fPIC -msse2 -mfpmath=sse -malign-double -ffast-math -fno-exceptions -fno-rtti -Wno-invalid-offsetof -Wno-unknown-pragmas -Wno-unused-result -g3 -gdwarf-2

LFLAGS  = -L/usr/local/PhysX-3.2.1-SDK/Lib/linux32 -Wl,--start-group -lPvdRuntimeCHECKED -lSimulationControllerCHECKED -lSceneQueryCHECKED -lLowLevelCHECKED -lLowLevelClothCHECKED -lPhysX3CHECKED -lPhysX3VehicleCHECKED -lPhysX3CookingCHECKED -lPhysX3ExtensionsCHECKED -lPhysX3CharacterKinematicCHECKED -lRepX3CHECKED -lRepXUpgrader3CHECKED -lPhysXProfileSDKCHECKED -lPhysXVisualDebuggerSDKCHECKED -lPxTaskCHECKED -lPhysX3CommonCHECKED -lGL -lGLU -lrt -lpthread -Wl,--end-group -m32

all: $(LIBRARY)

$(LIBRARY): $(SOURCES) *.h *.hpp
	g++ -shared $(SOURCES) $(CFLAGS) $(LFLAGS) -o $(LIBRARY)
	cp $(LIBRARY) $(INSTALLDIR)

clean:
	rm -f $(INSTALLDIR)/$(LIBRARY) $(LIBRARY)

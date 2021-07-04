package com.wingjoy.ghostshock.inter.myapplication;

import android.content.Context;
import android.provider.Settings;

public class AndroidId {
    public static String getAndroidId (Context context) {
        String ANDROID_ID = Settings.System.getString(context.getContentResolver(), Settings.System.ANDROID_ID);
        return ANDROID_ID;
    }
}

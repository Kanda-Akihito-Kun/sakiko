import { Call, type CancellablePromise } from "@wailsio/runtime";
import { Profile } from "../../bindings/sakiko.local/sakiko-core/interfaces";
import { SakikoService as GeneratedSakikoService } from "../../bindings/sakiko-wails";

export const SakikoService = {
  ...GeneratedSakikoService,
  DeleteResultArchive(taskID: string): CancellablePromise<void> {
    return Call.ByName("main.SakikoService.DeleteResultArchive", taskID);
  },
  SetProfileNodeEnabled(profileID: string, nodeIndex: number, enabled: boolean): CancellablePromise<Profile> {
    return Call.ByName("main.SakikoService.SetProfileNodeEnabled", profileID, nodeIndex, enabled).then((result: any) => {
      return Profile.createFrom(result);
    });
  },
};

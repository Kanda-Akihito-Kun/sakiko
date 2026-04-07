import { Call, type CancellablePromise } from "@wailsio/runtime";
import { SakikoService as GeneratedSakikoService } from "../../bindings/sakiko-wails";

export const SakikoService = {
  ...GeneratedSakikoService,
  DeleteResultArchive(taskID: string): CancellablePromise<void> {
    return Call.ByName("main.SakikoService.DeleteResultArchive", taskID);
  },
};

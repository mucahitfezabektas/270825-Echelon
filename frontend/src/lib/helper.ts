<script lang="ts" >
  import dayjs from "dayjs";

function formatTime(epoch: number | undefined): string {
    return epoch ? dayjs(epoch).format("HH:mm") : "-";
}
</script>